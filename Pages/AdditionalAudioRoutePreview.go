package Pages

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/gen2brain/malgo"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Utilities/AudioAPI"
)

type routeAudioPreviewSource struct {
	audioAPI  string
	device    string
	process   string
	processID int
}

func routeAudioPreviewSourceFromRoute(route Settings.AdditionalAudioRoute) routeAudioPreviewSource {
	return routeAudioPreviewSource{
		audioAPI:  route.Audio_api,
		device:    route.Audio_input_device,
		process:   route.Audio_input_process,
		processID: route.Audio_input_process_id,
	}
}

type routeAudioPreviewRequest struct {
	generation uint64
	source     routeAudioPreviewSource
}

// routeAudioInputPreview owns a temporary capture stream while an additional
// audio source is being edited. It never changes the running route and is
// stopped with the dialog.
type routeAudioInputPreview struct {
	level      *widget.ProgressBar
	meter      *audioThresholdMeter
	updates    chan routeAudioPreviewRequest
	stop       chan struct{}
	stopOnce   sync.Once
	generation atomic.Uint64
}

func newRouteAudioInputPreview(energy int) *routeAudioInputPreview {
	level := widget.NewProgressBar()
	level.Max = audioMeterMax
	level.TextFormatter = func() string { return "" }
	preview := &routeAudioInputPreview{
		level:   level,
		updates: make(chan routeAudioPreviewRequest, 1),
		stop:    make(chan struct{}),
	}
	preview.meter = newAudioThresholdMeter(level, energy)
	go preview.run()
	return preview
}

func (p *routeAudioInputPreview) CanvasObject() fyne.CanvasObject {
	return p.meter.CanvasObject()
}

func (p *routeAudioInputPreview) SetThreshold(energy int) {
	p.meter.SetThreshold(energy)
}

func (p *routeAudioInputPreview) Update(route Settings.AdditionalAudioRoute) {
	request := routeAudioPreviewRequest{
		generation: p.generation.Add(1),
		source:     routeAudioPreviewSourceFromRoute(route),
	}
	select {
	case <-p.stop:
		return
	default:
	}

	select {
	case p.updates <- request:
		return
	default:
	}
	// Only the newest selection matters when a user changes devices quickly.
	select {
	case <-p.updates:
	default:
	}
	select {
	case p.updates <- request:
	case <-p.stop:
	}
}

func (p *routeAudioInputPreview) Stop() {
	p.stopOnce.Do(func() {
		p.generation.Add(1)
		close(p.stop)
	})
}

func (p *routeAudioInputPreview) setLevel(generation uint64, level float64) {
	fyne.Do(func() {
		if p.generation.Load() == generation {
			p.level.SetValue(level)
		}
	})
}

func (p *routeAudioInputPreview) run() {
	var endpointMeter *routeEndpointAudioMeter
	var applicationMeter *Utilities.ApplicationAudioMeter
	cleanup := func() {
		if endpointMeter != nil {
			endpointMeter.Stop()
			endpointMeter = nil
		}
		if applicationMeter != nil {
			applicationMeter.Stop()
			applicationMeter = nil
		}
	}
	defer func() {
		cleanup()
		fyne.Do(func() { p.level.SetValue(0) })
	}()

	for {
		select {
		case <-p.stop:
			return
		case request := <-p.updates:
		drainUpdates:
			for {
				select {
				case newer := <-p.updates:
					request = newer
				default:
					break drainUpdates
				}
			}

			cleanup()
			p.setLevel(request.generation, 0)
			if strings.TrimSpace(request.source.process) != "" {
				processID := uint32(0)
				if request.source.processID > 0 {
					processID = uint32(request.source.processID)
				}
				meter, err := Utilities.StartApplicationAudioMeter(
					processID,
					request.source.process,
					func(peak float32) {
						p.setLevel(request.generation, normalizedAudioMeterLevel(float64(peak)))
					},
				)
				if err != nil {
					fmt.Printf("Could not start additional audio source application preview: %v\n", err)
					continue
				}
				applicationMeter = meter
				continue
			}

			meter, err := startRouteEndpointAudioMeter(request.source, func(level float64) {
				p.setLevel(request.generation, level)
			})
			if err != nil {
				fmt.Printf("Could not start additional audio source preview: %v\n", err)
				continue
			}
			endpointMeter = meter
		}
	}
}

type routeEndpointAudioMeter struct {
	context *malgo.AllocatedContext
	device  *malgo.Device
}

func startRouteEndpointAudioMeter(source routeAudioPreviewSource, onLevel func(float64)) (*routeEndpointAudioMeter, error) {
	backend := AudioAPI.GetAudioBackendByName(source.audioAPI)
	context, err := malgo.InitContext([]malgo.Backend{backend.Backend}, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, err
	}
	meter := &routeEndpointAudioMeter{context: context}
	fail := func(err error) (*routeEndpointAudioMeter, error) {
		meter.Stop()
		return nil, err
	}

	deviceType := malgo.Capture
	var selectedDevice *malgo.DeviceInfo
	deviceName := strings.TrimSpace(source.device)
	if deviceName != "" && !strings.EqualFold(deviceName, "Default") {
		captureDevices, captureErr := context.Devices(malgo.Capture)
		if captureErr != nil {
			return fail(captureErr)
		}
		for index := range captureDevices {
			if strings.EqualFold(strings.TrimSpace(captureDevices[index].Name()), deviceName) {
				selectedDevice = &captureDevices[index]
				break
			}
		}
		if selectedDevice == nil {
			loopbackDevices, loopbackErr := context.Devices(malgo.Loopback)
			if loopbackErr == nil {
				for index := range loopbackDevices {
					candidate := strings.TrimSpace(loopbackDevices[index].Name()) + " [Loopback]"
					if strings.EqualFold(candidate, deviceName) {
						selectedDevice = &loopbackDevices[index]
						deviceType = malgo.Loopback
						break
					}
				}
			}
		}
		if selectedDevice == nil {
			return fail(fmt.Errorf("audio input device %q was not found", deviceName))
		}
	}

	deviceConfig := malgo.DefaultDeviceConfig(deviceType)
	deviceConfig.Capture.Format = malgo.FormatS32
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = 16000
	deviceConfig.Alsa.NoMMap = 1
	if selectedDevice != nil {
		deviceConfig.Capture.DeviceID = selectedDevice.ID.Pointer()
	}
	callbacks := malgo.DeviceCallbacks{
		Data: func(_ []byte, input []byte, _ uint32) {
			if len(input) > 0 {
				onLevel(s32AudioPeakMeterLevel(input))
			}
		},
	}
	meter.device, err = malgo.InitDevice(context.Context, deviceConfig, callbacks)
	if err != nil {
		return fail(err)
	}
	if err = meter.device.Start(); err != nil {
		return fail(err)
	}
	return meter, nil
}

func (m *routeEndpointAudioMeter) Stop() {
	if m == nil {
		return
	}
	if m.device != nil {
		if m.device.IsStarted() {
			_ = m.device.Stop()
		}
		m.device.Uninit()
		m.device = nil
	}
	if m.context != nil {
		_ = m.context.Uninit()
		m.context.Free()
		m.context = nil
	}
}
