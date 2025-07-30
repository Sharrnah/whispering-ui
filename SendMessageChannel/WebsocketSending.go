package SendMessageChannel

import (
	"sync"
	"time"
)

// sending Message

type SendMessageStruct struct {
	Type  string      `json:"type"`
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

var SendMessageChannel = make(chan SendMessageStruct)

var timerLock sync.Mutex
var debounceTimers = make(map[string]*time.Timer)

const debounceDuration = 500 * time.Millisecond

func (message SendMessageStruct) SendMessage() {
	SendMessageChannel <- message
}

// SendMessageDebounced debounces messages based on Type and Name combination
// duration is optional - if not provided or zero, uses the default debounceDuration
func (message SendMessageStruct) SendMessageDebounced(duration ...time.Duration) {
	timerLock.Lock()
	defer timerLock.Unlock()

	// Use default duration if none provided or if provided duration is zero
	actualDuration := debounceDuration
	if len(duration) > 0 && duration[0] > 0 {
		actualDuration = duration[0]
	}

	// Create a unique key based on Type and Name
	key := message.Type + "_" + message.Name

	// Check if there's an existing timer for this Type+Name combination and stop it
	if timer, exists := debounceTimers[key]; exists {
		timer.Stop()
	}

	// Set up a new timer that calls the actual message sending after actualDuration
	debounceTimers[key] = time.AfterFunc(actualDuration, func() {
		SendMessageChannel <- message

		// Clean up the timer from the map after execution
		timerLock.Lock()
		delete(debounceTimers, key)
		timerLock.Unlock()
	})
}
