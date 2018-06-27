package actions

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/google/go-github/github"
)

// TransformListen converts a Github event into an Eventgrid event and published the event to Eventgrid
func TransformListen(c buffalo.Context) error {
	request := c.Request()
	/*payload, err := github.ValidatePayload(request, []byte(os.Getenv("APPSETTING_X_HUB_SIGNATURE")))
	if err != nil {
		log.Printf("secret key is not correct: err=%s\n", err)
		return c.Error(http.StatusInternalServerError, err)
	}*/
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
	myDate := date.Time{Time: time.Now()}

	var myClient = eventgrid.New()
	myClient.Authorizer = autorest.NewEventGridKeyAuthorizer(os.Getenv("APPSETTING_TOPIC_KEY"))
	eventName := FormatEventName(github.WebHookType(request))
	c.Logger().Debug(github.WebHookType(request))
	c.Logger().Debug(eventName)

	eventType := fmt.Sprintf("Github.%s", eventName)

	myEvent := eventgrid.Event{
		EventType:       to.StringPtr(eventType),
		EventTime:       &myDate,
		ID:              to.StringPtr(os.Getenv("APPSETTING_ID")),
		Data:            event,
		Subject:         to.StringPtr("GitToGrid"),
		DataVersion:     to.StringPtr("1"),
		MetadataVersion: to.StringPtr("1"),
	}
	var events []eventgrid.Event
	events = append(events, myEvent)
	result, err := myClient.PublishEvents(request.Context(), os.Getenv("APPSETTING_TOPIC_HOSTNAME"), events)
	if err != nil {
		log.Printf("Could not publish %s event to event grid: err=%s\n", eventName, err)
		return c.Error(result.Response.StatusCode, err)
	}
	return c.Render(200, render.JSON([]eventgrid.Event{myEvent}))
}

func FormatEventName(name string) string {
	sepName := strings.Split(name, "_")
	for i, v := range sepName {
		sepName[i] = strings.Title(v)
	}
	sepName = append(sepName, "Event")
	return strings.Join(sepName, "")
}
