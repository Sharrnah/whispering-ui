package Websocket

import (
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Websocket/Messages"
)

type streamRevision struct {
	revision int64
	final    bool
}

type streamingTranscriptState struct {
	revisions map[[2]string]streamRevision
	pending   map[string]string
}

func (s *streamingTranscriptState) apply(message MessageStruct) (string, bool) {
	if s.revisions == nil {
		s.revisions = make(map[[2]string]streamRevision)
		s.pending = make(map[string]string)
	}
	source := message.AudioSourceID
	if source == "" {
		source = "main"
	}
	key := [2]string{source, message.StreamID}
	previous, exists := s.revisions[key]
	if exists && (previous.final || message.StreamRevision <= previous.revision) {
		return "", false
	}
	text := ""
	if !message.Final {
		if err := json.Unmarshal(message.Data, &text); err != nil {
			return "", false
		}
	}
	s.revisions[key] = streamRevision{message.StreamRevision, message.Final}
	if len(s.revisions) > 128 {
		for old := range s.revisions {
			if old != key {
				delete(s.revisions, old)
				break
			}
		}
	}
	if message.Final {
		delete(s.pending, source)
	} else {
		if source != "main" && message.AudioSourceName != "" {
			text = message.AudioSourceName + ": " + text
		}
		s.pending[source] = text
	}
	sources := make([]string, 0, len(s.pending))
	for source := range s.pending {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	texts := make([]string, 0, len(sources))
	for _, source := range sources {
		texts = append(texts, s.pending[source])
	}
	return strings.Join(texts, "\n"), true
}

var streamedTranscripts streamingTranscriptState
var streamedHistory streamingTranscriptState
var streamedTranscriptMutex sync.Mutex

func handleStreamingTranscript(message MessageStruct) {
	streamedTranscriptMutex.Lock()
	defer streamedTranscriptMutex.Unlock()
	if message.DisplayMode == "blocks" && message.Type != "streaming_caption" {
		if _, accepted := streamedHistory.apply(message); !accepted {
			return
		}
		// History is immediate; the separate caption channel owns the preview
		// and can continue paging after recognition has finished.
		if message.Type == "transcript" && message.Final && strings.TrimSpace(message.Text) != "" {
			Messages.WhisperResult{Text: message.Text, Language: message.Language,
				TxtTranslation: message.TxtTranslation, TxtTranslationTarget: message.TxtTranslationTarget,
				AudioSourceID: message.AudioSourceID, AudioSourceName: message.AudioSourceName}.Update()
		}
		return
	}
	if message.Type == "streaming_caption" {
		message.StreamID = "caption:" + message.StreamID
		message.StreamRevision = message.DisplayRevision
		message.Final = message.DisplayDone
	}
	liveText, accepted := streamedTranscripts.apply(message)
	if !accepted {
		return
	}
	// Post UI updates in receive order. A goroutine per draft could run after
	// the final and restore an obsolete preview.
	if message.Type == "transcript" && message.Final && strings.TrimSpace(message.Text) != "" {
		Messages.WhisperResult{Text: message.Text, Language: message.Language,
			TxtTranslation: message.TxtTranslation, TxtTranslationTarget: message.TxtTranslationTarget,
			AudioSourceID: message.AudioSourceID, AudioSourceName: message.AudioSourceName}.Update()
	}
	realtimeLabelTimerMutex.Lock()
	if realtimeLabelTimer != nil {
		realtimeLabelTimer.Stop()
	}
	realtimeLabelTimerMutex.Unlock()
	fyne.Do(func() {
		Fields.DataBindings.WhisperResultIntermediateResult.Set(liveText)
		if liveText == "" {
			Fields.Field.ProcessingStatus.Stop()
		} else {
			Fields.Field.ProcessingStatus.Start()
			Fields.Field.RealtimeResultScroll.Show()
		}
	})
}
