package loggly

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gliderlabs/logspout/router"
)

const (
	adapterName         = "loggly"
	logglyTokenEnvVar   = "LOGGLY_TOKEN"
	logglyTagsEnvVar    = "LOGGLY_TAGS"
	logglyTagsHeader    = "X-LOGGLY-TAG"
	logglyAddr          = "https://logs-01.loggly.com"
	logglyEventEndpoint = "/inputs"
)

func init() {
	router.AdapterFactories.Register(NewLogglyAdapter, adapterName)

	r := &router.Route{
		Adapter: "loggly",
	}

	// It's not documented in the logspout repo but if you want to use an adapter
	// without going through the routesapi you must add at #init or via #New...
	err := router.Routes.Add(r)
	if err != nil {
		log.Fatal("could not add route: ", err.Error())
	}
}

// NewLogglyAdapter returns an Adapter with that uses a loggly token taken from
// the LOGGLY_TOKEN environment variable
func NewLogglyAdapter(route *router.Route) (router.LogAdapter, error) {
	token := os.Getenv(logglyTokenEnvVar)

	if token == "" {
		return nil, errors.New("could not find environment variable LOGGLY_TOKEN")
	}

	return &Adapter{
		token: token,
		client: http.Client{
			Timeout: 900 * time.Millisecond, // logspout will cull any spout that does  respond within 1 second
		},
		tags: os.Getenv(logglyTagsEnvVar),
		log:  log.New(os.Stdout, "logspout-loggly", log.LstdFlags),
	}, nil
}

// Adapter satisfies the router.LogAdapter interface by providing Stream which
// passes all messages to loggly.
type Adapter struct {
	token  string
	client http.Client
	tags   string
	log    *log.Logger
}

// Stream satisfies the router.LogAdapter interface and passes all messages to
// Loggly
func (l *Adapter) Stream(logstream chan *router.Message) {
	for m := range logstream {
		msg := logglyMessage{
			Message:           m.Data,
			ContainerName:     m.Container.Name,
			ContainerID:       m.Container.ID,
			ContainerImage:    m.Container.Config.Image,
			ContainerHostname: m.Container.Config.Hostname,
		}

		l.SendMessage(msg)
	}
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
	defer resp.Body.Close()

	if err != nil {
		l.log.Println(
			fmt.Errorf(
				"error from client: %s",
				err.Error(),
			),
		)
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

type logglyMessage struct {
	Message           string `json:"message"`
	ContainerName     string `json:"container_name"`
	ContainerID       string `json:"container_id"`
	ContainerImage    string `json:"container_image"`
	ContainerHostname string `json:"hostname"`
}
