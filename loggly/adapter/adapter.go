package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gliderlabs/logspout/router"
)

const (
	logglyAddr          = "https://logs-01.loggly.com"
	logglyEventEndpoint = "/bulk"
	flushTimeout        = 10 * time.Second
)

// Adapter satisfies the router.LogAdapter interface by providing Stream which
// passes all messages to loggly.
type Adapter struct {
	bufferSize int
	log        *log.Logger
	logglyURL  string
	queue      chan logglyMessage
}

// New returns an Adapter that receives messages from logspout. Additionally,
// it launches a goroutine to buffer and flush messages to loggly.
func New(logglyToken string, tags string, bufferSize int) *Adapter {
	adapter := &Adapter{
		bufferSize: bufferSize,
		log:        log.New(os.Stdout, "logspout-loggly", log.LstdFlags),
		logglyURL:  buildLogglyURL(logglyToken, tags),
		queue:      make(chan logglyMessage),
	}

	go adapter.readQueue()

	return adapter
}

// Stream satisfies the router.LogAdapter interface and passes all messages to
// Loggly
func (l *Adapter) Stream(logstream chan *router.Message) {
	for m := range logstream {
		l.queue <- logglyMessage{
			Message:           m.Data,
			ContainerName:     m.Container.Name,
			ContainerID:       m.Container.ID,
			ContainerImage:    m.Container.Config.Image,
			ContainerHostname: m.Container.Config.Hostname,
		}
	}
}

func (l *Adapter) readQueue() {
	buffer := l.newBuffer()

	timeout := time.NewTimer(flushTimeout)

	for {
		select {
		case msg := <-l.queue:
			if len(buffer) == cap(buffer) {
				timeout.Stop()
				l.flushBuffer(buffer)
				buffer = l.newBuffer()
			}

			buffer = append(buffer, msg)

		case <-timeout.C:
			if len(buffer) > 0 {
				l.flushBuffer(buffer)
				buffer = l.newBuffer()
			}
		}

		timeout.Reset(flushTimeout)
	}
}

func (l *Adapter) newBuffer() []logglyMessage {
	return make([]logglyMessage, 0, l.bufferSize)
}

func (l *Adapter) flushBuffer(buffer []logglyMessage) {
	var data bytes.Buffer

	for _, msg := range buffer {
		j, _ := json.Marshal(msg)
		data.Write(j)
		data.WriteString("\n")
	}

	req, _ := http.NewRequest(
		"POST",
		l.logglyURL,
		&data,
	)

	go l.sendRequestToLoggly(req)
}

func (l *Adapter) sendRequestToLoggly(req *http.Request) {
	resp, err := http.DefaultClient.Do(req)

	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		l.log.Println(
			fmt.Errorf(
				"error from client: %s",
				err.Error(),
			),
		)
		return
	}

	if resp.StatusCode != http.StatusOK {
		l.log.Println(
			fmt.Errorf(
				"received a %s status code when sending message. response: %s",
				resp.StatusCode,
				resp.Body,
			),
		)
	}
}

func buildLogglyURL(token, tags string) string {
	var url string
	url = fmt.Sprintf(
		"%s%s/%s",
		logglyAddr,
		logglyEventEndpoint,
		token,
	)

	if tags != "" {
		url = fmt.Sprintf(
			"%s/tag/%s/",
			url,
			tags,
		)
	}
	return url
}
