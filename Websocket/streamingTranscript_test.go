package Websocket

import (
	"encoding/json"
	"testing"
)

func TestStreamingTranscriptOrderingAndSourceIsolation(t *testing.T) {
	state := streamingTranscriptState{}
	message := MessageStruct{Streaming: true, StreamID: "utterance", StreamRevision: 1,
		AudioSourceID: "main", Type: "processing_data", Data: json.RawMessage(`"hello"`)}
	if text, ok := state.apply(message); !ok || text != "hello" {
		t.Fatalf("first draft = %q, %v", text, ok)
	}
	game := message
	game.AudioSourceID, game.AudioSourceName = "game", "Game"
	if text, ok := state.apply(game); !ok || text != "Game: hello\nhello" {
		t.Fatalf("independent draft = %q, %v", text, ok)
	}
	message.Final, message.StreamRevision = true, 2
	if text, ok := state.apply(message); !ok || text != "Game: hello" {
		t.Fatalf("final cleared another source: %q, %v", text, ok)
	}
	message.Final, message.StreamRevision = false, 3
	if _, ok := state.apply(message); ok {
		t.Fatal("accepted a draft after final")
	}
	message.StreamID = "next-utterance"
	if _, ok := state.apply(message); !ok {
		t.Fatal("rejected a new utterance")
	}
}
