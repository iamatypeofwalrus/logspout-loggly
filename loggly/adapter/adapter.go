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
	logglyTagsHeader    = "X-LOGGLY-TAG"
	flushTimeout        = 10 * time.Second
)

// Adapter satisfies the router.LogAdapter interface by providing Stream which
// passes all messages to loggly.
type Adapter struct {
	token      string
	client     *http.Client
	tags       string
	log        *log.Logger
	queue      chan logglyMessage
	bufferSize int
}

// New returns an Adapter that receives messages from logspout. Additionally,
// it launches a goroutine to buffer and flush messages to loggly.
func New(logglyToken, tags string, bufferSize int) *Adapter {

	adapter := &Adapter{
		client:     http.DefaultClient,
		bufferSize: bufferSize,
		log:        log.New(os.Stdout, "logspout-loggly", log.LstdFlags),
		queue:      make(chan logglyMessage),
		token:      logglyToken,
		tags:       tags,
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
		fmt.Sprintf("%s%s/%s", logglyAddr, logglyEventEndpoint, l.token),
		&data,
	)

	go l.sendRequestToLoggly(req)
}

func (l *Adapter) sendRequestToLoggly(req *http.Request) {
	resp, err := l.client.Do(req)
	defer resp.Body.Close()

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
				"received a non 200 status code when sending message to loggly: %s",
				err.Error(),
			),
		)
	}
}
