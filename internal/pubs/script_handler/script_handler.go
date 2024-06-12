package script_handler

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"os/exec"
	"strings"
	"time"
)

type ScriptHandler struct {
	count        int
	command      string
	chanHandlers chan interface{}
}

func (r *ScriptHandler) Publish(data interface{}) error {
	r.chanHandlers <- data
	return nil
}

func NewScriptHandler(command string, count int, queueSize int) *ScriptHandler {
	handlers := &ScriptHandler{
		count:        count,
		command:      command,
		chanHandlers: make(chan interface{}, queueSize),
	}

	for i := 0; i < count; i++ {
		go func() {
			handlers.worker(i + 1)
		}()
	}

	go func() {
		for {
			time.Sleep(time.Second * 30)
			logrus.Infof("queue size: %d, in queue %v", queueSize, len(handlers.chanHandlers))
		}
	}()

	return handlers
}

func (h *ScriptHandler) worker(num int) {
	logrus.Infof("Starting worker num %v", num)
	for {
		data := <-h.chanHandlers
		jsonData, err := json.Marshal(data)
		if err != nil {
			logrus.Errorf("error marshalling to JSON: %s", err)
			continue
		}

		// Команда для запуска скрипта
		cmd := exec.Command("/bin/bash", "-c", h.command)

		// Буфер для передачи данных на стандартный ввод
		cmd.Stdin = bytes.NewBuffer(jsonData)

		// Буферы для захвата stdout и stderr
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		// Выполнение команды
		err = cmd.Run()
		if err != nil {
			logrus.Errorf("error executing script: %v", err)
		}
		if strings.TrimSpace(stdout.String()) != "" {
			logrus.Infof("script stdout: %v", stdout.String())
		}
		if strings.TrimSpace(stderr.String()) != "" {
			logrus.Errorf("script stderr: %v", stderr.String())
		}

	}
}
