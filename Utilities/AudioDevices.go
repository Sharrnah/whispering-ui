package Utilities

import (
	"fmt"
	"github.com/gen2brain/malgo"
	"whispering-tiger-ui/CustomWidget"
)

type AudioDevice struct {
	Name      string
	Index     int
	ID        string
	IsDefault bool
}

type AudioDeviceList struct {
	DeviceType malgo.DeviceType
	Devices    []AudioDevice
}

type AudioDeviceMemory struct {
	Backend       malgo.Backend
	Devices       []AudioDevice
	WidgetOptions []CustomWidget.TextValueOption
}

// WASAPI device lists
var AudioInputDevicesListWASAPI AudioDeviceMemory
var AudioOutputDevicesListWASAPI AudioDeviceMemory

// MME device lists
var AudioInputDevicesListMME AudioDeviceMemory
var AudioOutputDevicesListMME AudioDeviceMemory

// DirectSound device lists
var AudioInputDevicesListDirectSound AudioDeviceMemory
var AudioOutputDevicesListDirectSound AudioDeviceMemory

func InitMalgo(audioAPI malgo.Backend) (*malgo.AllocatedContext, error) {
	// initialize malgo
	var backends = []malgo.Backend{audioAPI}

	ctx, err := malgo.InitContext(backends, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	if err != nil {
		fmt.Println(err)
		//os.Exit(1)
		return nil, err
	}
	return ctx, nil
}

func GetAudioDevices(audioAPI malgo.Backend, deviceType malgo.DeviceType, deviceIndexStartPoint int) ([]AudioDevice, error) {
	//a.DeviceType = deviceType
	ctx, err := InitMalgo(audioAPI)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	devices, err := ctx.Devices(deviceType)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var nameSuffix string
	// if deviceType is Loopback, change it to Capture to get the correct device info (and prevent crash of miniaudio)
	if deviceType == malgo.Loopback {
		deviceType = malgo.Capture
		nameSuffix = " [Loopback]"
	}

	deviceList := make([]AudioDevice, 0)
	for index, deviceInfo := range devices {
		fullInfo, err := ctx.DeviceInfo(deviceType, deviceInfo.ID, malgo.Shared)
		if err != nil {
			continue
		}
		deviceList = append(deviceList, AudioDevice{
			Name:      deviceInfo.Name() + nameSuffix,
			Index:     index + deviceIndexStartPoint,
			ID:        deviceInfo.ID.String(),
			IsDefault: fullInfo.IsDefault != 0,
		})
	}

	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	return deviceList, nil
}
