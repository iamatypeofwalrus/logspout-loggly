package loggly

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gliderlabs/logspout/router"
)

const (
	logglyTokenEnvVar   = "LOGGLY_TOKEN"
	logglyAddr          = "https://logs-01.loggly.com"
	logglyEventEndpoint = "/inputs"
)

func init() {
	router.AdapterFactories.Register(NewLogglyAdapter, "loggly")
}

// NewLogglyAdapter returns an Adapter with that uses a loggly token taken from
// the LOGGLY_TOKEN environment variable
func NewLogglyAdapter(route *router.Route) (router.LogAdapter, error) {
	token := os.Getenv(logglyTokenEnvVar)

	if token == "" {
		log.Fatal("Could not find environment variable LOGGLY_TOKEN")
		return nil, errors.New("")
	}

	log.Println("Creating Loggly Adapter")

	return &Adapter{
		token:  token,
		client: http.Client{},
	}, nil
}

// Adapter satisfies the router.LogAdapter interface by providing Stream which
// passes all messages on to loggly.
type Adapter struct {
	token  string
	client http.Client
}

// Stream satisfies the router.LogAdapter interface and passes all logs to Loggly
func (l *Adapter) Stream(logstream chan *router.Message) {
	for m := range logstream {
		log.Println("Received message from stream")
		msg := logglyMessage{
			Message:           m.Data,
			ContainerName:     m.Container.Name,
			ContainerID:       m.Container.ID,
			ContainerImage:    m.Container.Config.Image,
			ContainerHostname: m.Container.Config.Hostname,
		}

		js, err := json.Marshal(msg)

		if err != nil {
			log.Fatal(err.Error())
		}

		url := fmt.Sprintf("%s%s", logglyAddr, logglyEventEndpoint)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(js))

		if err != nil {
			log.Fatal(err.Error())
		}

		log.Println("Sending request to loggly")
		_, err = l.client.Do(req)

		if err != nil {
			errMsg := fmt.Sprintf("Error from client: %s", err.Error())
			log.Fatal(errMsg)
		}

	}
}

type logglyMessage struct {
	Message           string `json:"message"`
	ContainerName     string `json:"container_name"`
	ContainerID       string `json:"container_id"`
	ContainerImage    string `json:"container_image"`
	ContainerHostname string `json:"hostname"`
}
