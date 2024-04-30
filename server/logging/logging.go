package logging

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
	"util"
)

var key []byte

var client *http.Client

func init() {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
}

func SetKey(serverKey []byte) {
	key = serverKey
}

func SendLogRemote(action string) {
	currentTime := time.Now().Format("2006/01/02 15:04:05")
	logMessage := fmt.Sprintf("%s INFO %s", currentTime, action)

	req, err := http.NewRequest("POST", "https://localhost:10444/logs", bytes.NewReader([]byte(logMessage)))
	req.Header.Set("Authorization", util.Encode64(key))
	util.FailOnError(err)
	client.Do(req)
}
