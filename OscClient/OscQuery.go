package OscClient

import (
	"encoding/json"
	zeroconf "github.com/lunarhue/metallic-flock-zeroconf"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type oscNodeDescriptor struct {
	FullPath string `json:"FULL_PATH"`
}

type oscRootResponse struct {
	FullPath string                       `json:"FULL_PATH"`
	Contents map[string]oscNodeDescriptor `json:"CONTENTS"`
}

type hostInfoResponse struct {
	OscPort int `json:"OSC_PORT"`
}

func RunOscQuery() {
	// 1) OSC UDP server on a random free port (Port:0)
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 0, // let OS choose
	})
	if err != nil {
		log.Fatalf("ListenUDP failed: %v", err)
	}
	defer udpConn.Close()

	oscPort := udpConn.LocalAddr().(*net.UDPAddr).Port
	log.Printf("OSC UDP listening on %s (port %d)", udpConn.LocalAddr(), oscPort)

	// Dummy reader: you would parse OSC here
	go func() {
		buf := make([]byte, 2048)
		for {
			n, src, readErr := udpConn.ReadFromUDP(buf)
			if readErr != nil {
				log.Printf("ReadFromUDP error: %v", readErr)
				return
			}
			log.Printf("Received %d bytes over OSC from %s", n, src)
		}
	}()

	// 2) HTTP server for OSCQuery, on a random free TCP port
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// VRChat calls "/?HOST_INFO" to get the OSC UDP port
		if strings.EqualFold(r.URL.RawQuery, "HOST_INFO") {
			resp := hostInfoResponse{OscPort: oscPort}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				log.Printf("Encode HOST_INFO failed: %v", err)
			}
			return
		}

		// Root node: show what paths we "support"
		root := oscRootResponse{
			FullPath: "/",
			Contents: map[string]oscNodeDescriptor{
				"avatar": {
					FullPath: "/avatar",
				},
				"tracking": {
					FullPath: "/tracking",
				},
			},
		}

		if err := json.NewEncoder(w).Encode(root); err != nil {
			log.Printf("Encode root failed: %v", err)
		}
	})

	// Listen on TCP port 0 to get an available port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("TCP Listen failed: %v", err)
	}
	httpPort := ln.Addr().(*net.TCPAddr).Port
	log.Printf("OSCQuery HTTP server on http://%s (port %d)", ln.Addr().String(), httpPort)

	// Start HTTP server using that listener
	httpServer := &http.Server{Handler: mux}
	go func() {
		if err := httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 3) Zeroconf advertise the HTTP (OSCQuery) server
	serviceType := zeroconf.NewType("_oscjson._tcp")
	serviceName := "WhisperingTigerOSCApp"

	service := zeroconf.NewService(serviceType, serviceName, uint16(httpPort))
	client, err := zeroconf.New().Publish(service).Open()
	if err != nil {
		log.Fatalf("zeroconf publish failed: %v", err)
	}
	defer client.Close()

	log.Printf("Published Zeroconf service %q of type %q on port %d",
		serviceName, serviceType, httpPort)

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	_ = httpServer.Close()
}
