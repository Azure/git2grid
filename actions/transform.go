package actions

import (
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

// TransformListen converts a Github event into an Eventgrid event and publishes the event to Eventgrid
func TransformListen(c buffalo.Context) error {

	request := c.Request()
	// payload, err := github.ValidatePayload(request, []byte(os.Getenv("APPSETTING_X_HUB_SIGNATURE")))
	// if err != nil {
	// 	log.Printf("secret key is not correct: err=%s\n", err)
	// 	return c.Error(http.StatusInternalServerError, err)
	// }
	var err error
	var body []byte
	if body, err = ioutil.ReadAll(request.Body); err != nil {
		return err
	}
	defer request.Body.Close()
	event, err := github.ParseWebHook(github.WebHookType(request), body)

	if err != nil {
		log.Printf("Could not parse webhook: err=%s\n", err)
		return c.Error(http.StatusInternalServerError, err)
	}

	var myClient = eventgrid.New()
	myClient.Authorizer = autorest.NewEventGridKeyAuthorizer(os.Getenv("eventgridSecret"))

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

	result, err := myClient.PublishEvents(request.Context(), "gridfromgit.westus2-1.eventgrid.azure.net", events)

	if err != nil {
		log.Printf("Could not publish %s event to event grid: err=%s\n", github.WebHookType(request), err)
		return c.Error(result.Response.StatusCode, err)
	}

	return c.Render(200, render.JSON([]eventgrid.Event{myEvent}))
}
