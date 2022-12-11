package RuntimeBackend

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"whispering-tiger-ui/Utilities"
)

var BackendsList []WhisperProcessConfig

// RunWithStreams ComposeWithStreams executes a command
// stdin/stdout/stderr
func (c *WhisperProcessConfig) RunWithStreams(name string, arguments []string, stdIn io.Reader, stdOut io.Writer, stdErr io.Writer, action ...string) error {
	var arg []string

	for _, file := range arguments {
		arg = append(arg, file)
	}

	arg = append(arg, action...)

	proc := exec.Command(name, arg...)
	proc.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	// attach environment variables
	if c.environmentVars != nil {
		proc.Env = c.environmentVars
	}

	proc.Stdout = stdOut
	proc.Stdin = stdIn
	proc.Stderr = stdErr

	c.Program = proc

	return proc.Run()
}

type WhisperProcessConfig struct {
	DeviceIndex       string
	DeviceOutIndex    string
	SettingsFile      string
	Program           *exec.Cmd
	ReaderBackend     *io.PipeReader
	WriterBackend     *io.PipeWriter
	environmentVars []string
}

func NewWhisperProcess() WhisperProcessConfig {
	var ReaderBackend, WriterBackend = io.Pipe()

	return WhisperProcessConfig{
		DeviceIndex:    "-1",
		DeviceOutIndex: "-1",
		SettingsFile:   "settings.yaml",
		ReaderBackend:  ReaderBackend,
		WriterBackend:  WriterBackend,
	}
}

func (c *WhisperProcessConfig) IsRunning() bool {
	if c.Program.Process == nil {
		return false
	}

	return true
}

func (c *WhisperProcessConfig) Stop() {
	if c.Program != nil && c.Program.Process != nil {
		println("Terminating process")
		_ = c.Program.Process.Signal(syscall.SIGINT)
		_ = c.Program.Process.Signal(syscall.SIGKILL)
		_ = c.Program.Process.Signal(syscall.SIGTERM)
		//_ = c.Program.Process.Kill()

		c.Program.Stdout = nil
		c.Program.Stdin = nil
		c.Program.Stderr = nil
	}
}

func (c *WhisperProcessConfig) AttachEnvironment(envName, envValue string) {
	c.environmentVars = os.Environ()

	envIndex := -1
	for index, element := range c.environmentVars {
		if strings.HasPrefix(element, envName + "=") {
			envIndex = index
		}
	}

	if value, ok := os.LookupEnv(envName); !ok {
		c.environmentVars = append(c.environmentVars, envName + "=" + envValue)
	} else {
		if envIndex > -1 {
			c.environmentVars[envIndex] = envName + "=" + envValue + ";" + value
		} else {
			c.environmentVars = append(c.environmentVars, envName + "=" + envValue + ";" + value)
		}
	}
}

func (c *WhisperProcessConfig) Start() {
	go func(writer io.Writer, reader io.Reader) {
		var tmpReader io.Reader
		var err error

		if Utilities.FileExists("audioWhisper.py") {
			err = c.RunWithStreams("python", []string{"-u", "audioWhisper.py",
				"--device_index", c.DeviceIndex,
				"--device_out_index", c.DeviceOutIndex,
				"--config", c.SettingsFile,
			}, tmpReader, writer, writer)
		} else if Utilities.FileExists("audioWhisper/audioWhisper.exe") {
			err = c.RunWithStreams("audioWhisper/audioWhisper.exe", []string{
				"--device_index", c.DeviceIndex,
				"--device_out_index", c.DeviceOutIndex,
				"--config", c.SettingsFile,
			}, tmpReader, writer, writer)
		} else {
			err = errors.New("could not start audioWhisper")
		}

		if err != nil {
			_, _ = writer.Write([]byte("Error: " + err.Error()))
			return
		}
	}(c.WriterBackend, c.ReaderBackend)
}
