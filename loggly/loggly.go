package loggly

import (
	"errors"
	"log"
	"os"

	"github.com/gliderlabs/logspout/router"
	"github.com/iamatypeofwalrus/logspout-loggly/loggly/adapter"
)

const (
	adapterName       = "loggly"
	filterNameEnvVar  = "FILTER_NAME"
	logglyTokenEnvVar = "LOGGLY_TOKEN"
	logglyTagsEnvVar  = "LOGGLY_TAGS"
)

func init() {
	router.AdapterFactories.Register(NewLogglyAdapter, adapterName)

	r := &router.Route{
		Adapter:    "loggly",
		FilterName: os.Getenv(filterNameEnvVar),
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
	tags := os.Getenv(logglyTagsEnvVar)

	if token == "" {
		return nil, errors.New(
			"could not find environment variable LOGGLY_TOKEN",
		)
	}

	return adapter.New(
		token,
		tags,
		100,
	), nil
}
