package webhooklog

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"zktoro/config"

	"zktoro/zktoro-core-go/clients/webhook/client/operations"

	"github.com/goccy/go-json"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// Logger logs the alerts to a log file.
type Logger struct {
	file *os.File
}

// NewLogger creates a new logger.
func NewLogger(logFileName string) (*Logger, error) {
	logsDir := path.Join(config.DefaultContainerzktoroDirPath, "logs")
	if err := os.MkdirAll(logsDir, 0777); err != nil {
		return nil, fmt.Errorf("failed to create the logs dir: %v", err)
	}

	fileName := fmt.Sprintf("zktoro-local-alerts-logs-%d", time.Now().Unix())
	if len(logFileName) > 0 {
		fileName = logFileName
	}

	fullPath := path.Join(logsDir, fileName)
	log.WithField("file", fullPath).Info("logging local alerts and metrics")

	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create the webhook log file: %v", err)
	}
	log.WithField("path", fullPath).Info("logging webhook alerts")
	go func() {
		<-context.Background().Done()
		file.Close()
	}()
	return &Logger{file: file}, nil
}

// Close implemenets io.Closer.
func (logger *Logger) Close() error {
	if logger.file != nil {
		return logger.file.Close()
	}
	return nil
}

// SendAlerts logs the webhook alert to a line-delimited file by marshalling to JSON.
func (logger *Logger) SendAlerts(params *operations.SendAlertsParams, opts ...operations.ClientOption) (*operations.SendAlertsOK, error) {
	b, _ := json.Marshal(params.Payload)
	_, err := fmt.Fprintln(logger.file, string(b))
	if err != nil {
		return nil, fmt.Errorf("failed to write the webhook alert log: %v", err)
	}
	return &operations.SendAlertsOK{}, nil
}

// StdoutLogger logs the alerts to stdout.
type StdoutLogger struct{}

// NewStdoutLogger creates a new logger.
func NewStdoutLogger() (*StdoutLogger, error) {
	return &StdoutLogger{}, nil
}

// Close implemenets io.Closer.
func (logger *StdoutLogger) Close() error {
	return nil
}

// SendAlerts logs the webhook alert to a line-delimited file by marshalling to JSON.
func (logger *StdoutLogger) SendAlerts(params *operations.SendAlertsParams, opts ...operations.ClientOption) (*operations.SendAlertsOK, error) {
	b, _ := json.Marshal(params.Payload)
	logrus.Info(string(b))
	return &operations.SendAlertsOK{}, nil
}
