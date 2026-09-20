package RemoteAudio

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net"
	"sort"
	"strconv"
	"time"
)

type Host struct {
	Name    string `json:"name"`
	Address string `json:"-"`
	Port    int    `json:"port"`
	Service string `json:"service"`
	Version int    `json:"version"`
	Nonce   string `json:"nonce"`
}

func Discover(timeout time.Duration) ([]Host, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero})
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	nonceBytes := make([]byte, 16)
	if _, err = rand.Read(nonceBytes); err != nil {
		return nil, err
	}
	nonce := hex.EncodeToString(nonceBytes)
	request := []byte("WT_AUDIO_DISCOVER " + nonce)
	targets := []net.IP{net.IPv4bcast, net.IPv4(127, 0, 0, 1)}
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addresses, _ := iface.Addrs()
		for _, address := range addresses {
			ip, subnet, err := net.ParseCIDR(address.String())
			if err != nil || ip.To4() == nil {
				continue
			}
			ip = ip.To4()
			mask := subnet.Mask
			if len(mask) != 4 {
				continue
			}
			broadcast := make(net.IP, 4)
			for i := 0; i < 4; i++ {
				broadcast[i] = ip[i] | ^mask[i]
			}
			targets = append(targets, broadcast)
		}
	}
	for _, ip := range targets {
		_, _ = conn.WriteToUDP(request, &net.UDPAddr{IP: ip, Port: 5001})
	}
	conn.SetReadDeadline(time.Now().Add(timeout))
	found := map[string]Host{}
	buffer := make([]byte, 1024)
	for {
		n, peer, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			return nil, err
		}
		var host Host
		if json.Unmarshal(buffer[:n], &host) != nil || host.Service != "whispering-tiger-audio" || host.Version != 1 || host.Nonce != nonce || host.Port < 1 || host.Port > 65535 {
			continue
		}
		// Use the packet's source IP; never trust a discovery-advertised URL.
		host.Address = "ws://" + net.JoinHostPort(peer.IP.String(), strconv.Itoa(host.Port))
		found[host.Address] = host
	}
	result := make([]Host, 0, len(found))
	for _, host := range found {
		result = append(result, host)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Address < result[j].Address })
	return result, nil
}
