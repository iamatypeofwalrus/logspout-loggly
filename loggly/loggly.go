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
	adapterName         = "loggly"
	logglyTokenEnvVar   = "LOGGLY_TOKEN"
	logglyTagsEnvVar    = "LOGGLY_TAGS"
	logglyAddr          = "https://logs-01.loggly.com"
	logglyEventEndpoint = "/inputs"
)

// TODO: consider logging all fatals to loggly

func init() {
	router.AdapterFactories.Register(NewLogglyAdapter, adapterName)

	r := &router.Route{
		Adapter: "loggly",
	}

	// It's not documented in the logspout repo but if you want to use an adapter without
	// going through the routesapi you must add at #init or via #New...
	err := router.Routes.Add(r)
	if err != nil {
		log.Fatal("Could not add route: ", err.Error())
	}
}

// NewLogglyAdapter returns an Adapter with that uses a loggly token taken from
// the LOGGLY_TOKEN environment variable
func NewLogglyAdapter(route *router.Route) (router.LogAdapter, error) {
	token := os.Getenv(logglyTokenEnvVar)

	if token == "" {
		return nil, errors.New("Could not find environment variable LOGGLY_TOKEN")
	}

	return &Adapter{
		token:  token,
		client: http.Client{},
	}, nil
}

// Adapter satisfies the router.LogAdapter interface by providing Stream which
// passes all messages to loggly.
type Adapter struct {
	token  string
	client http.Client
}

// Stream satisfies the router.LogAdapter interface and passes all messages to Loggly
func (l *Adapter) Stream(logstream chan *router.Message) {
	for m := range logstream {
		msg := logglyMessage{
			Message:           m.Data,
			ContainerName:     m.Container.Name,
			ContainerID:       m.Container.ID,
			ContainerImage:    m.Container.Config.Image,
			ContainerHostname: m.Container.Config.Hostname,
		}

		err := l.SendMessage(msg)

		if err != nil {
			log.Println(err.Error())
		}
	}
}

// SendMessage handles creating and sending a request to Loggly. Any errors that occur during that
// process are bubbled up to the caller
func (l *Adapter) SendMessage(msg logglyMessage) error {
	js, err := json.Marshal(msg)

	if err != nil {
		return err
	}

  fmt.Fprint(os.Stdout, logglyTagsEnvVar)
	url := fmt.Sprintf("%s%s/%s", logglyAddr, logglyEventEndpoint, l.token)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(js))

	req.Header.Add("X-LOGGLY-TAG", os.Getenv(logglyTagsEnvVar))
	fmt.Fprint(os.Stdout, req.Header)

	if err != nil {
		return err
	}

	// TODO: possibly use pool of workers to send requests?
	resp, err := l.client.Do(req)

	if err != nil {
		errMsg := fmt.Sprintf("Error from client: %s", err.Error())
		return errors.New(errMsg)
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Received a non 200 status code: %s", err.Error())
		return errors.New(errMsg)
	}

	return nil
}

type logglyMessage struct {
	Message           string `json:"message"`
	ContainerName     string `json:"container_name"`
	ContainerID       string `json:"container_id"`
	ContainerImage    string `json:"container_image"`
	ContainerHostname string `json:"hostname"`
}
