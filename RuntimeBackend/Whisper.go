package RuntimeBackend

import (
	"encoding/json"
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	Utilities.ProcessHideWindowAttr(proc)

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
	go c.SetOutputHandling(stdErrPipeReader, c.processErrorOutputLine)

	go c.SetOutputHandling(stdOutPipeReader, c.processLogOutputLine)

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
		SettingsFile:   filepath.Join(".", "Profiles", "settings.yaml"),
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

func (c *WhisperProcessConfig) processLogOutputLine(line string, isUpdating bool) {
	// Try to decode if the line contains a progress percentage
	progress, err := Utilities.ParseProgressFromString(line)
	if err == nil {
		Fields.Field.StatusBar.SetValue(progress)
	} else {
		Fields.Field.StatusBar.SetValue(0)
	}

	if !isUpdating {
		// Try to decode the line as loading JSON message
		if ProcessLoadingMessage(line) {
			// if it is a loading message, do not add to log
			return
		}

		// write to log.txt file if WriteLogFile is enabled
		if fyne.CurrentApp().Preferences().BoolWithFallback("WriteLogfile", false) {
			Utilities.WriteLog("log.txt", line)
		}

		// write to c.RecentLog
		c.RecentLog = append(c.RecentLog, line)
		// remove excess elements if length is greater than 2000
		if len(c.RecentLog) > MaxClipboardLogLines {
			c.RecentLog = c.RecentLog[len(c.RecentLog)-MaxClipboardLogLines:]
		}
	}

	// set last log line to status bar text.
	Fields.DataBindings.StatusTextBinding.Set(line)
}

func (c *WhisperProcessConfig) processErrorOutputLine(line string, isUpdating bool) {
	// Try to decode if the line contains a progress percentage
	progress, err := Utilities.ParseProgressFromString(line)
	if err == nil {
		Fields.Field.StatusBar.SetValue(progress)
	} else {
		Fields.Field.StatusBar.SetValue(0)
	}

	if !isUpdating {
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
			currentMainWindow, _ := Utilities.GetCurrentMainWindow("")
			dialog.ShowError(errors.New(exceptionMessage.Error+"\n\n"+lastTraceback), currentMainWindow)
		}

		// Try to decode the line as loading JSON message
		if ProcessLoadingMessage(line) {
			// if it is a loading message, do not add to log
			return
		}

		// write to log.txt file if WriteLogFile is enabled
		if fyne.CurrentApp().Preferences().BoolWithFallback("WriteLogfile", false) {
			Utilities.WriteLog("log.txt", line)
		}

		// write to c.RecentLog
		c.RecentLog = append(c.RecentLog, line)
		// remove excess elements if length is greater than 2000
		if len(c.RecentLog) > MaxClipboardLogLines {
			c.RecentLog = c.RecentLog[len(c.RecentLog)-MaxClipboardLogLines:]
		}
	}

	// set last log line to status bar text.
	Fields.DataBindings.StatusTextBinding.Set(line)
}

func (c *WhisperProcessConfig) SetOutputHandling(stderr io.Reader, processLineFunc func(string, bool)) {
	bufLen := 4096
	buf := make([]byte, bufLen)
	var incompleteLine string

	for {
		if stderr == nil {
			break
		}
		num, err := stderr.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				if len(incompleteLine) > 0 {
					processLineFunc(incompleteLine, false)
				}
				break
			} else if err, ok := err.(*os.PathError); ok && err.Err.Error() == "input/output error" {
				log.Printf("read error: %v", err)
				break
			}
		}

		output := string(buf[:num])
		lines := strings.Split(output, "\n")

		if len(incompleteLine) > 0 {
			lines[0] = incompleteLine + lines[0]
			incompleteLine = ""
		}

		for i, line := range lines {
			// Handle carriage return (updating lines)
			if strings.Contains(line, "\r") {
				parts := strings.Split(line, "\r")
				for j, part := range parts {
					if j > 0 {
						if i == len(lines)-1 && j == len(parts)-1 {
							incompleteLine = part
						}
						if strings.TrimSpace(part) != "" {
							processLineFunc(part, true) // Treat as updated line
						}
					} else {
						if i == len(lines)-1 && output[len(output)-1] != '\n' {
							incompleteLine = part
						}
						if strings.TrimSpace(part) != "" {
							processLineFunc(part, false)
						}
					}
				}
			} else if i == len(lines)-1 && output[len(output)-1] != '\n' {
				incompleteLine = line
				if strings.TrimSpace(line) != "" {
					processLineFunc(line, true) // line might be updated
				}
			} else {
				if strings.TrimSpace(line) != "" {
					processLineFunc(line, false)
				}
			}
		}
	}

	// Ensure any remaining incomplete line is processed after the loop
	if len(incompleteLine) > 0 {
		if strings.TrimSpace(incompleteLine) != "" {
			processLineFunc(incompleteLine, false)
		}
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
		} else if Utilities.FileExists("audioWhisper/audioWhisper") { // Linux variant without file extension
			err = c.RunWithStreams("audioWhisper/audioWhisper", cmdArguments, tmpReader, c.WriterBackend, c.WriterBackend)
		} else if Utilities.FileExists("audioWhisper/audioWhisper.py") && Utilities.FileExists("audioWhisper/venv/Scripts/python.exe") {
			c.AttachEnvironment("VIRTUAL_ENV", "audioWhisper/venv/")
			cmdArguments = append([]string{"-u", "audioWhisper.py"}, cmdArguments...)
			err = c.RunWithStreams("audioWhisper/venv/Scripts/python.exe", cmdArguments, tmpReader, c.WriterBackend, c.WriterBackend)
		} else if Utilities.FileExists("audioWhisper/audioWhisper.py") && Utilities.FileExists("audioWhisper/venv/Scripts/python") { // Linux variant without file extension
			c.AttachEnvironment("VIRTUAL_ENV", "audioWhisper/venv/")
			cmdArguments = append([]string{"-u", "audioWhisper.py"}, cmdArguments...)
			err = c.RunWithStreams("audioWhisper/venv/Scripts/python", cmdArguments, tmpReader, c.WriterBackend, c.WriterBackend)
		} else {
			err = errors.New("could not start audioWhisper")
		}

		if err != nil {
			_, _ = c.WriterBackend.Write([]byte("Error: " + err.Error()))
			return
		}
	}(stdoutTee)
}

func RestartBackend(confirmation bool, confirmationText string) {
	currentMainWindow, isNewWindow := Utilities.GetCurrentMainWindow(lang.L("Restart Backend"))

	restartFunction := func() {
		// close running backend process
		if len(BackendsList) > 0 && BackendsList[0].IsRunning() {
			infinityProcessDialog := dialog.NewCustom(lang.L("Restarting Backend"), lang.L("OK"), container.NewVBox(widget.NewLabel(lang.L("Restarting Backend")+"..."), widget.NewProgressBarInfinite()), currentMainWindow)
			infinityProcessDialog.Show()
			infinityProcessDialog.SetOnClosed(func() {
				if isNewWindow {
					currentMainWindow.Close()
				}
			})
			BackendsList[0].Stop()
			time.Sleep(2 * time.Second)
			BackendsList[0].Start()
			infinityProcessDialog.Hide()
			Fields.DataBindings.SpeechToTextEnabledDataBinding.Set(true)
		}
	}
	if confirmation {
		confirmationDialog := dialog.NewConfirm(lang.L("Restart Backend"), confirmationText, func(b bool) {
			if b {
				restartFunction()
			}
		}, currentMainWindow)
		confirmationDialog.Show()
	} else {
		restartFunction()
	}
}
