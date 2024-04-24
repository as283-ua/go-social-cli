package logging

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"util"
)

const key = "clave_secreta"

var defaultLogger *slog.Logger
var client *http.Client

func getLogger() *slog.Logger {
	return slog.New(slog.Default().Handler())
}

func init() {
	defaultLogger = getLogger()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
}

func Error(msg string) {
	defaultLogger.Error(msg)
}

func Errorf(msg string, a ...any) {
	defaultLogger.Error(fmt.Sprintf(msg, a...))
}

func Info(msg string) {
	defaultLogger.Info(msg)
}

func Warn(msg string) {
	defaultLogger.Warn(msg)
}

func SendLogRemote(action string) {
	currentTime := time.Now().Format("2006/01/02 15:04:05")
	logMessage := fmt.Sprintf("%s INFO %s", currentTime, action)

	req, err := http.NewRequest("POST", "https://localhost:10444/logs", bytes.NewReader([]byte(logMessage)))
	req.Header.Set("Authorization", key)
	util.FailOnError(err)
	client.Do(req)
}
