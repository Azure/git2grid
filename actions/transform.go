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
)

// TransformListen default implementation.
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
		//Should a request.Body.Close() go here?
		return err
	}
	defer request.Body.Close()
	event, err := github.ParseWebHook(github.WebHookType(request), body)
	if err != nil {
		log.Printf("could not parse webhook: err=%s\n", err)
		return c.Error(http.StatusInternalServerError, err)
	}
	//c.Logger().Debug("MADE IT")
	//repoName := ""
	var myEvent eventgrid.Event
	//can take address of myDate directly because it is a local variable
	myDate := date.Time{Time: time.Now()}
	var events []eventgrid.Event
	var myClient = eventgrid.New()
	myClient.Authorizer = autorest.NewEventGridKeyAuthorizer(os.Getenv("APPSETTING_TOPIC_KEY"))
	switch e := event.(type) {
	case *github.PullRequestEvent:
		if e.Action != nil {
			//repoName = *e.Repo.FullName
			myEvent = eventgrid.Event{
				EventType:       to.StringPtr(os.Getenv("APPSETTING_EVENT_TYPE")),
				EventTime:       &myDate,
				ID:              to.StringPtr(os.Getenv("APPSETTING_ID")),
				Data:            e,
				Subject:         e.PullRequest.URL,
				DataVersion:     to.StringPtr("1"),
				MetadataVersion: to.StringPtr("1"),
			}
			events = append(events, myEvent)
			result, err := myClient.PublishEvents(request.Context(), os.Getenv("APPSETTING_TOPIC_HOSTNAME"), events)
			if err != nil {
				log.Printf("Could not publish pull request event to event grid: err=%s\n", err)
				return c.Error(result.Response.StatusCode, err)
			}
		}
	case *github.LabelEvent:
		if e.Action != nil {
			//repoName = *e.Repo.FullName
			myEvent = eventgrid.Event{
				EventType:       to.StringPtr(os.Getenv("APPSETTING_EVENT_TYPE")),
				EventTime:       &myDate,
				ID:              to.StringPtr(os.Getenv("APPSETTING_ID")),
				Data:            e,
				Subject:         e.Label.URL,
				DataVersion:     to.StringPtr("11"),
				MetadataVersion: to.StringPtr("1"),
			}
			events = append(events, myEvent)
			result, err := myClient.PublishEvents(request.Context(), os.Getenv("APPSETTING_TOPIC_HOSTNAME"), events)
			if err != nil {
				log.Printf("Could not publish label event to event grid: err=%s\n", err)
				return c.Error(result.Response.StatusCode, err)
			}
		}
	default:
		log.Printf("unknown event type %s\n", github.WebHookType(request))
		//return err
	}
	return c.Render(200, render.JSON([]eventgrid.Event{myEvent}))
}
