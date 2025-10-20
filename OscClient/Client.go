package OscClient

import (
	"github.com/hypebeast/go-osc/osc"
	"whispering-tiger-ui/Settings"
)

func createClient() *osc.Client {
	ip := Settings.Config.Osc_ip
	port := Settings.Config.Osc_port
	return osc.NewClient(ip, port)
}

func SendBool(address string, value bool) {
	client := createClient()
	msg := osc.NewMessage(address)
	msg.Append(value)
	_ = client.Send(msg)
}
