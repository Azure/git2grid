package actions

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/envy"
	"github.com/google/go-github/github"
	"github.com/satori/go.uuid"
)

// GithookReceiver converts a Github event into an Eventgrid event and publishes the event to Eventgrid
func GithookReceiver(c buffalo.Context) error {
	webhookSecret := envy.Get("webhookSecret", "")
	eventgridSecret := envy.Get("eventgridSecret", "")
	eventgridEndpoint := envy.Get("eventgridEndpoint", "")

	request := c.Request()

	// save request body so it can be read multiple times
	var bodyBytes []byte

	if request.Body != nil {
		// max expected body is 25 megabytes
		var thirtyMegaBytes int64 = 30 * 1024 * 1024
		limitReader := io.LimitReader(request.Body, thirtyMegaBytes)
		bodyBytes, _ = ioutil.ReadAll(limitReader)
	}

	defer request.Body.Close()
	request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	// validate githook using secret
	_, err := github.ValidatePayload(request, []byte(webhookSecret))

	if err != nil {
		c.Logger().Errorf("validation failed: err=%s\n", err)
		return c.Error(http.StatusInternalServerError, err)
	}

	// parse webhook body
	event, err := github.ParseWebHook(github.WebHookType(request), bodyBytes)

	if err != nil {
		c.Logger().Errorf("could not parse webhook body: err=%s\n", err)
		return c.Error(http.StatusInternalServerError, err)
	}

	// publish to eventgrid
	var eventgridClient = eventgrid.New()
	eventgridClient.Authorizer = autorest.NewEventGridKeyAuthorizer(eventgridSecret)

	var now = date.Time{Time: time.Now()}
	uuidBytes := uuid.NewV4()
	var uuidString = string(uuidBytes[:16])

	var subject = fmt.Sprintf("/webhook/%s", github.WebHookType(request))

	myEvent := eventgrid.Event{
		EventType:       to.StringPtr(github.WebHookType(request)),
		EventTime:       &now,
		ID:              to.StringPtr(uuidString),
		Data:            event,
		Subject:         to.StringPtr(subject),
		DataVersion:     to.StringPtr("1"),
		MetadataVersion: to.StringPtr("1"),
	}

	var events []eventgrid.Event
	events = append(events, myEvent)

	result, err := eventgridClient.PublishEvents(request.Context(), eventgridEndpoint, events)

	if err != nil {
		c.Logger().Errorf("could not publish %s event to event grid: err=%s\n", github.WebHookType(request), err)
		return c.Error(result.Response.StatusCode, err)
	}

	return c.Render(200, render.JSON([]eventgrid.Event{myEvent}))
}

func FormatSubjectName(name string) string {
	sepName := strings.Split(name, "_")
	for i, v := range sepName {
		sepName[i] = strings.Title(v)
	}
	sepName = append(sepName, "Event")
	return strings.Join(sepName, "")
}
