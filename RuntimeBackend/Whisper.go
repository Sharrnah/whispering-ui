package RuntimeBackend

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Utilities"
)

var BackendsList []WhisperProcessConfig

const MaxClipboardLogLines = 2000

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

	// Create a new pipe for the StdErr field
	stdErrPipeReader, stdErrPipeWriter := io.Pipe()
	// Create a new pipe for the StdOut field
	stdOutPipeReader, stdOutPipeWriter := io.Pipe()

	//proc.Stdout = stdOut
	proc.Stdout = io.MultiWriter(stdOut, stdOutPipeWriter)
	proc.Stdin = stdIn
	proc.Stderr = io.MultiWriter(stdErr, stdErrPipeWriter)

	c.Program = proc

	// parse for errors coming from the backend process and printed to stderr
	go c.SetErrorOutputHandling(stdErrPipeReader)

	go c.SetLogOutputHandling(stdOutPipeReader)

	return proc.Run()
}

type WhisperProcessConfig struct {
	DeviceIndex     string
	DeviceOutIndex  string
	SettingsFile    string
	UiDownload      bool
	Program         *exec.Cmd
	ReaderBackend   *io.PipeReader
	WriterBackend   *io.PipeWriter
	environmentVars []string
	RecentLog       []string
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
	timeout := 6 * time.Second

	if c.Program != nil && c.Program.Process != nil {
		println("Terminating process")

		sendMessage := Fields.SendMessageStruct{
			Type:  "quit",
			Name:  "quit",
			Value: "",
		}
		sendMessage.SendMessage()

		time.Sleep(timeout)

		if c.Program.Process != nil {
			_ = c.Program.Process.Signal(syscall.SIGINT)
			time.Sleep(timeout / 2)
		}

		if c.Program.Process != nil {
			_ = c.Program.Process.Signal(syscall.SIGKILL)
			_ = c.Program.Process.Signal(syscall.SIGTERM)
		}

		c.Program.Stdout = nil
		c.Program.Stdin = nil
		c.Program.Stderr = nil

		c.Program.Process = nil
	}
}

func (c *WhisperProcessConfig) AttachEnvironment(envName, envValue string) {
	c.environmentVars = os.Environ()

	envIndex := -1
	for index, element := range c.environmentVars {
		if strings.HasPrefix(element, envName+"=") {
			envIndex = index
		}
	}

	if value, ok := os.LookupEnv(envName); !ok {
		c.environmentVars = append(c.environmentVars, envName+"="+envValue)
	} else {
		if envIndex > -1 {
			c.environmentVars[envIndex] = envName + "=" + envValue + ";" + value
		} else {
			c.environmentVars = append(c.environmentVars, envName+"="+envValue+";"+value)
		}
	}
}

func (c *WhisperProcessConfig) SetLogOutputHandling(stdout io.Reader) {
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		// write to c.RecentLog
		c.RecentLog = append(c.RecentLog, line)
		// remove first element if length is greater than 2000
		if len(c.RecentLog) > MaxClipboardLogLines {
			c.RecentLog = c.RecentLog[1:]
		}
	}
}

func (c *WhisperProcessConfig) SetErrorOutputHandling(stdout io.Reader) {
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		// write to log.txt file if WriteLogFile is enabled
		if fyne.CurrentApp().Preferences().BoolWithFallback("WriteLogfile", false) {
			Utilities.WriteLog("log.txt", line)
		}

		// Try to decode the line as a JSON message
		var exceptionMessage struct {
			Type      string   `json:"type"`
			Error     string   `json:"message"`
			Traceback []string `json:"traceback"`
		}
		if err := json.Unmarshal([]byte(line), &exceptionMessage); err == nil {
			lastTraceback := ""
			if exceptionMessage.Traceback != nil && len(exceptionMessage.Traceback) > 0 {
				lastTraceback = exceptionMessage.Traceback[len(exceptionMessage.Traceback)-1]
			}
			// Handle error message
			if len(fyne.CurrentApp().Driver().AllWindows()) == 1 && fyne.CurrentApp().Driver().AllWindows()[0] != nil {
				dialog.ShowError(errors.New(exceptionMessage.Error+"\n\n"+lastTraceback), fyne.CurrentApp().Driver().AllWindows()[0])
			} else if len(fyne.CurrentApp().Driver().AllWindows()) == 2 && fyne.CurrentApp().Driver().AllWindows()[1] != nil {
				dialog.ShowError(errors.New(exceptionMessage.Error+"\n\n"+lastTraceback), fyne.CurrentApp().Driver().AllWindows()[1])
			} else {
				fmt.Printf("%s\n", exceptionMessage.Error)
			}
		}

		// write to c.RecentLog
		c.RecentLog = append(c.RecentLog, line)
		// remove first element if length is greater than 2000
		if len(c.RecentLog) > MaxClipboardLogLines {
			c.RecentLog = c.RecentLog[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		log.Print(err)
	}
}

func (c *WhisperProcessConfig) Start() {
	defer Utilities.PanicLogger()

	// Create a pipe to capture the output from the process
	_, pw := io.Pipe()

	// Create a multi-writer to write to both c.WriterBackend and &stdOutBuf
	multiWriter := io.MultiWriter(c.WriterBackend, pw)

	// Create a tee reader to duplicate the output from the process
	stdoutTee := io.TeeReader(c.ReaderBackend, multiWriter)

	go func(stdOut io.Reader) {
		defer Utilities.PanicLogger()

		var tmpReader io.Reader
		var err error

		cmdArguments := []string{
			"--device_index", c.DeviceIndex,
			"--device_out_index", c.DeviceOutIndex,
			"--config", c.SettingsFile,
		}

		if c.UiDownload {
			cmdArguments = append(cmdArguments, "--ui_download")
		}

		if Utilities.FileExists("audioWhisper.py") {
			cmdArguments = append([]string{"-u", "audioWhisper.py"}, cmdArguments...)
			err = c.RunWithStreams("python", cmdArguments, tmpReader, c.WriterBackend, c.WriterBackend)
		} else if Utilities.FileExists("audioWhisper/audioWhisper.exe") {
			err = c.RunWithStreams("audioWhisper/audioWhisper.exe", cmdArguments, tmpReader, c.WriterBackend, c.WriterBackend)
		} else {
			err = errors.New("could not start audioWhisper")
		}

		if err != nil {
			_, _ = c.WriterBackend.Write([]byte("Error: " + err.Error()))
			return
		}
	}(stdoutTee)
}
