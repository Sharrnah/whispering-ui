package Updater

import (
	"errors"
	"sync"
)

var (
	ErrNoDownloadURLs        = errors.New("no download URLs were provided")
	ErrDownloadSessionActive = errors.New("download session is already running")
)

// DownloadSession coordinates one resumable file download across a list of
// mirrors. Pausing or changing the mirror cancels only the current HTTP
// requests; the next attempt keeps the already written prefix on disk.
type DownloadSession struct {
	URLs                []string
	Filepath            string
	ConcurrentDownloads int
	ChunkSize           int64

	OnAttempt  func(downloadURL string, serverIndex, serverCount int)
	OnProgress func(downloadURL string, progress, total uint64, speed float64, resuming bool)
	OnPaused   func()
	OnResumed  func()

	stateMu            sync.Mutex
	stateChanged       *sync.Cond
	activeDownloader   *Download
	currentServerIndex int
	running            bool
	pauseRequested     bool
	paused             bool
	switchRequested    bool
}

func NewDownloadSession(urls []string, filepath string, concurrentDownloads int, chunkSize int64) *DownloadSession {
	session := &DownloadSession{
		URLs:                uniqueNonEmptyURLs(urls),
		Filepath:            filepath,
		ConcurrentDownloads: concurrentDownloads,
		ChunkSize:           chunkSize,
	}
	session.stateChanged = sync.NewCond(&session.stateMu)
	return session
}

func uniqueNonEmptyURLs(urls []string) []string {
	uniqueURLs := make([]string, 0, len(urls))
	seen := make(map[string]struct{}, len(urls))
	for _, downloadURL := range urls {
		if downloadURL == "" {
			continue
		}
		if _, exists := seen[downloadURL]; exists {
			continue
		}
		seen[downloadURL] = struct{}{}
		uniqueURLs = append(uniqueURLs, downloadURL)
	}
	return uniqueURLs
}

func (s *DownloadSession) ServerCount() int {
	return len(s.URLs)
}

func (s *DownloadSession) IsPaused() bool {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	return s.paused
}

// Pause requests a resumable pause. It returns false if no transfer is active
// or another control action is already being processed.
func (s *DownloadSession) Pause() bool {
	s.stateMu.Lock()
	if !s.running || s.activeDownloader == nil || s.pauseRequested || s.paused || s.switchRequested {
		s.stateMu.Unlock()
		return false
	}
	s.pauseRequested = true
	downloader := s.activeDownloader
	s.stateMu.Unlock()

	downloader.Cancel()
	return true
}

// Resume releases a session that has reached its paused state.
func (s *DownloadSession) Resume() bool {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if !s.running || !s.paused {
		return false
	}
	s.paused = false
	s.pauseRequested = false
	s.stateChanged.Broadcast()
	return true
}

// SwitchServer cancels the current requests and continues from the next mirror.
func (s *DownloadSession) SwitchServer() bool {
	s.stateMu.Lock()
	if len(s.URLs) <= 1 || !s.running || s.activeDownloader == nil || s.pauseRequested || s.paused || s.switchRequested {
		s.stateMu.Unlock()
		return false
	}
	s.switchRequested = true
	downloader := s.activeDownloader
	s.stateMu.Unlock()

	downloader.Cancel()
	return true
}

func (s *DownloadSession) advanceServer() {
	s.stateMu.Lock()
	s.currentServerIndex = (s.currentServerIndex + 1) % len(s.URLs)
	s.stateMu.Unlock()
}

// Run downloads the file and automatically tries every mirror once before
// returning the final error. Run blocks while paused and must therefore be
// called from a worker goroutine when used by a UI.
func (s *DownloadSession) Run(retries int) error {
	if len(s.URLs) == 0 {
		return ErrNoDownloadURLs
	}

	s.stateMu.Lock()
	if s.running {
		s.stateMu.Unlock()
		return ErrDownloadSessionActive
	}
	s.running = true
	s.currentServerIndex = 0
	s.pauseRequested = false
	s.paused = false
	s.switchRequested = false
	s.stateMu.Unlock()

	defer func() {
		s.stateMu.Lock()
		s.activeDownloader = nil
		s.running = false
		s.pauseRequested = false
		s.paused = false
		s.switchRequested = false
		s.stateChanged.Broadcast()
		s.stateMu.Unlock()
	}()

	automaticFailures := 0
	for {
		s.stateMu.Lock()
		serverIndex := s.currentServerIndex
		downloadURL := s.URLs[serverIndex]
		downloader := &Download{
			Url:                 downloadURL,
			Filepath:            s.Filepath,
			ConcurrentDownloads: s.ConcurrentDownloads,
			ChunkSize:           s.ChunkSize,
		}
		downloader.WriteCounter.OnProgress = func(progress, total uint64, speed float64) {
			if s.OnProgress != nil {
				s.OnProgress(downloadURL, progress, total, speed, downloader.IsResuming())
			}
		}
		s.activeDownloader = downloader
		s.stateMu.Unlock()

		if s.OnAttempt != nil {
			s.OnAttempt(downloadURL, serverIndex, len(s.URLs))
		}

		err := downloader.DownloadFile(retries)

		s.stateMu.Lock()
		wasPaused := s.pauseRequested
		wasSwitchRequested := s.switchRequested
		s.activeDownloader = nil
		s.switchRequested = false
		if err == nil {
			s.pauseRequested = false
			s.paused = false
			s.stateChanged.Broadcast()
		}
		s.stateMu.Unlock()

		if err == nil {
			return nil
		}

		// A requested control action wins over a simultaneous transport error.
		// The partial file remains available for the next attempt.
		if wasSwitchRequested {
			automaticFailures = 0
			s.advanceServer()
			continue
		}
		if wasPaused {
			s.stateMu.Lock()
			s.paused = true
			s.stateMu.Unlock()
			if s.OnPaused != nil {
				s.OnPaused()
			}

			s.stateMu.Lock()
			for s.paused {
				s.stateChanged.Wait()
			}
			s.stateMu.Unlock()
			if s.OnResumed != nil {
				s.OnResumed()
			}
			continue
		}

		if errors.Is(err, ErrDownloadCanceled) {
			return err
		}

		automaticFailures++
		if automaticFailures >= len(s.URLs) {
			return err
		}
		s.advanceServer()
	}
}
