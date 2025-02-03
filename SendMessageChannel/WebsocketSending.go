package SendMessageChannel

// sending Message

type SendMessageStruct struct {
	Type  string      `json:"type"`
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

var SendMessageChannel = make(chan SendMessageStruct)

func (message SendMessageStruct) SendMessage() {
	SendMessageChannel <- message
}
