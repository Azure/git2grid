package server

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/google/go-github/github"
	"github.com/google/uuid"
)

// GithookReceiver converts a Github event into an Eventgrid event and publishes the event to Eventgrid
func GithookReceiver(c buffalo.Context) error {
	webhookSecret := os.Getenv("webhookSecret")
	eventgridSecret := os.Getenv("eventgridSecret")
	eventgridEndpoint := os.Getenv("eventgridEndpoint")

	request := c.Request()

	// save request body so it can be read multiple times
	var bodyBytes []byte

	if request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(request.Body)
	}

	request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	// validate githook using secret
	_, err := github.ValidatePayload(request, []byte(webhookSecret))

	if err != nil {
		log.Printf("Validation failed: err=%s\n", err)
		return c.Error(http.StatusInternalServerError, err)
	}

	// parse webhook body
	defer request.Body.Close()
	event, err := github.ParseWebHook(github.WebHookType(request), bodyBytes)

	if err != nil {
		log.Printf("Could not parse webhook body: err=%s\n", err)
		return c.Error(http.StatusInternalServerError, err)
	}

	// publish to eventgrid
	var eventgridClient = eventgrid.New()
	eventgridClient.Authorizer = autorest.NewEventGridKeyAuthorizer(eventgridSecret)

	var now = date.Time{Time: time.Now()}
	var uuidBytes, _ = uuid.NewRandom()
	var uuidString = string(uuidBytes[:16])

	myEvent := eventgrid.Event{
		EventType:       to.StringPtr(github.WebHookType(request)),
		EventTime:       &now,
		ID:              to.StringPtr(uuidString),
		Data:            event,
		Subject:         to.StringPtr("GitToGrid"),
		DataVersion:     to.StringPtr("1"),
		MetadataVersion: to.StringPtr("1"),
	}

	var events []eventgrid.Event
	events = append(events, myEvent)

	result, err := eventgridClient.PublishEvents(request.Context(), eventgridEndpoint, events)

	if err != nil {
		log.Printf("Could not publish %s event to event grid: err=%s\n", github.WebHookType(request), err)
		return c.Error(result.Response.StatusCode, err)
	}

	return c.Render(200, render.JSON([]eventgrid.Event{myEvent}))
}
