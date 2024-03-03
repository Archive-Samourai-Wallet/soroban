package ipc

import (
	"bufio"
	"context"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func StartProcessDaemon(ctx context.Context, name, process string, args ...string) {
	for {
		select {
		case <-ctx.Done():
		default:
			log.Trace("Starting process")

			<-time.After(3 * time.Second)
			startSubProcess(ctx, name, process, args...)
		}
	}
}

func startSubProcess(ctx context.Context, name, process string, args ...string) {
	cmd := exec.Command(process, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.WithError(err).Fatal("Failed to get process stdout")
	}
	cmd.Stderr = cmd.Stdout
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		log.WithError(err).Fatal("Failed to start sub process")
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		text := strings.Trim(scanner.Text(), "\n\r ")
		if len(text) == 0 {
			continue
		}
		log.Debugf("%d - %s: %s", cmd.Process.Pid, name, text)
	}
	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		log.WithError(err).
			Fatal("Command finished with error")
	}
}
