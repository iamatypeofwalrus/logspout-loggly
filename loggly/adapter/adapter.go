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

// New returns an Adapter that receives messages from logspout and launches
// a goroutine to buffer and flush messages to loggly.
func New(logglyToken, tags string, bufferSize int) *Adapter {
	client := &http.Client{
		Timeout: 900 * time.Millisecond, // logspout will cull any spout that does  respond within 1 second
	}

	adapter := &Adapter{
		client:     client,
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

	for msg := range l.queue {
		if len(buffer) == cap(buffer) {
			l.flushBuffer(buffer)
			buffer = l.newBuffer()
		}

		buffer = append(buffer, msg)
	}
}

func (l *Adapter) newBuffer() []logglyMessage {
	return make([]logglyMessage, 0, l.bufferSize)
}

func (l *Adapter) flushBuffer(buffer []logglyMessage) {

}

// SendMessage handles creating and sending a request to Loggly. Any errors
// that occur during that process are bubbled up to the caller
func (l *Adapter) SendMessage(msg logglyMessage) {
	js, err := json.Marshal(msg)

	if err != nil {
		log.Println(err)
		return
	}

	url := fmt.Sprintf("%s%s/%s", logglyAddr, logglyEventEndpoint, l.token)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(js))

	if err != nil {
		l.log.Println(err)
		return
	}

	if l.tags != "" {
		req.Header.Add(logglyTagsHeader, l.tags)
	}

	go l.sendRequestToLoggly(req)
}

func (l *Adapter) sendRequestToLoggly(req *http.Request) {
	resp, err := l.client.Do(req)

	if err != nil {
		l.log.Println(
			fmt.Errorf(
				"error from client: %s",
				err.Error(),
			),
		)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		l.log.Println(
			fmt.Errorf(
				"received a non 200 status code when sending message to loggly: %s",
				err.Error(),
			),
		)
	}
}
