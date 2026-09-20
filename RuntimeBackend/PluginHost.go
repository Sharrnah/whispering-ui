package RuntimeBackend

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
	"whispering-tiger-ui/Utilities"
)

// PluginHost owns a private child process. It never sends lifecycle commands
// to the AI PC's shared backend/control connection.
type PluginHost struct {
	cmd      *exec.Cmd
	input    io.WriteCloser
	outgoing chan []byte
	done     chan struct{}
	once     sync.Once
	jobOnce  sync.Once
	owner    WhisperProcessConfig
}

func pluginHostCommand(config string) (*exec.Cmd, error) {
	args := []string{"--plugin_host", "--config", config}
	for _, binary := range []string{"audioWhisper/audioWhisper.exe", "audioWhisper/audioWhisper"} {
		if Utilities.FileExists(binary) {
			return exec.Command(binary, args...), nil
		}
	}
	for _, root := range []string{".", "audioWhisper"} {
		script := filepath.Join(root, "audioWhisper.py")
		if !Utilities.FileExists(script) {
			continue
		}
		python := "python"
		for _, candidate := range []string{"venv/Scripts/python.exe", "venv/bin/python", "venv/Scripts/python"} {
			path := filepath.Join(root, candidate)
			if Utilities.FileExists(path) {
				python = path
				break
			}
		}
		return exec.Command(python, append([]string{"-u", script}, args...)...), nil
	}
	return nil, errors.New("local plugin host requires the updated audioWhisper backend bundle")
}

func StartPluginHost(config string, onMessage func([]byte), onExit func(error)) (*PluginHost, error) {
	absolute, err := filepath.Abs(config)
	if err != nil {
		return nil, err
	}
	cmd, err := pluginHostCommand(absolute)
	if err != nil {
		return nil, err
	}
	return startPluginHostCommand(cmd, onMessage, onExit)
}

func startPluginHostCommand(cmd *exec.Cmd, onMessage func([]byte), onExit func(error)) (*PluginHost, error) {
	Utilities.ProcessHideWindowAttr(cmd)
	setNewProcessGroup(cmd)
	input, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	output, err := cmd.StdoutPipe()
	if err != nil {
		input.Close()
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		input.Close()
		output.Close()
		return nil, err
	}
	host := &PluginHost{cmd: cmd, input: input, outgoing: make(chan []byte, 64), done: make(chan struct{})}
	if err = cmd.Start(); err != nil {
		input.Close()
		output.Close()
		stderr.Close()
		return nil, err
	}
	if err = host.owner.assignProcessToJobObject(cmd.Process.Pid); err != nil {
		input.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		host.releaseJob()
		return nil, err
	}
	go func() {
		defer input.Close()
		for {
			select {
			case <-host.done:
				return
			case raw := <-host.outgoing:
				if _, err := input.Write(append(raw, '\n')); err != nil {
					return
				}
			}
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Buffer(make([]byte, 4096), 1024*1024)
		for scanner.Scan() {
			log.Printf("Local plugins: %s", scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(output)
		scanner.Buffer(make([]byte, 4096), 1024*1024)
		for scanner.Scan() {
			onMessage(append([]byte(nil), scanner.Bytes()...))
		}
		scanErr := scanner.Err()
		if scanErr != nil {
			host.releaseJob()
			killPluginProcessTree(cmd)
		}
		err := cmd.Wait()
		host.releaseJob()
		close(host.done)
		if scanErr != nil {
			err = scanErr
		}
		onExit(err)
	}()
	return host, nil
}

func (h *PluginHost) Send(value interface{}) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if len(raw) > 1024*1024 {
		return fmt.Errorf("local plugin message too large")
	}
	select {
	case <-h.done:
		return os.ErrClosed
	default:
	}
	select {
	case h.outgoing <- raw:
		return nil
	default:
		return errors.New("local plugin queue is full")
	}
}

func (h *PluginHost) Close() {
	h.once.Do(func() {
		// EOF is a graceful shutdown even if a plugin filled the command queue.
		_ = h.input.Close()
		select {
		case <-h.done:
		case <-time.After(5 * time.Second):
			h.releaseJob()
			killPluginProcessTree(h.cmd)
			<-h.done
		}
	})
}

func (h *PluginHost) releaseJob() { h.jobOnce.Do(func() { closeJobObject(&h.owner) }) }
