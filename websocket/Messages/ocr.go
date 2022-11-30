package Messages

type WindowsStruct struct {
	Windows []string `json:"data"`
}

var WindowsList WindowsStruct

func (res WindowsStruct) Update() {
	//
}
