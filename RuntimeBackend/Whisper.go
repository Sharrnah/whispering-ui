package RuntimeBackend

import (
	"io"
	"os"
	"os/exec"
)

// RunWithStreams ComposeWithStreams executes a command
// stdin/stdout/stderr
func RunWithStreams(name string, arguments []string, stdIn io.Reader, stdOut io.Writer, stdErr io.Writer, action ...string) (*exec.Cmd, error) {
	var arg []string

	for _, file := range arguments {
		arg = append(arg, file)
	}

	arg = append(arg, action...)

	proc := exec.Command(name, arg...)
	//stdin, _ = proc.StdoutPipe()
	//stdout, _ = proc.StdinPipe()
	//stderr, _ = proc.StderrPipe()

	proc.Stdout = stdOut
	proc.Stdin = stdIn
	proc.Stderr = stdErr

	return proc, proc.Run()
	//return stdIn, stdOut, stdErr
}

//goland:noinspection GoExportedOwnDeclaration
var ReaderBackend, WriterBackend = io.Pipe()

type WhisperProcessConfig struct {
	DeviceIndex string
	SettingsFile     string
	Process *os.Process
}

func (c *WhisperProcessConfig) StartWhisper() {
	go func(writer io.Writer, reader io.Reader) {
		var tmpReader io.Reader
		//var tmpWriterReader io.Writer
		proc, err := RunWithStreams("python", []string{"-u", "audioWhisper.py", "--websocket_ip", "127.0.0.1", "--device_index", c.DeviceIndex, "--config", c.SettingsFile}, tmpReader, writer, writer)
		if err != nil {
			_, _ = writer.Write([]byte("Error: " + err.Error()))
		}
		c.Process = proc.Process
	}(WriterBackend, ReaderBackend)
}
