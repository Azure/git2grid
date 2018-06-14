package actions

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/eventgrid/2018-01-01/eventgrid"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/google/go-github/github"
)

// HomeHandler is a default handler to serve up
// a home page.
func HomeHandler(c buffalo.Context) error {
	request := c.Request()
	payload, err := github.ValidatePayload(request, []byte(os.Getenv("APPSETTING_X_HUB_SIGNATURE")))
	if err != nil {
		log.Printf("secret key is not correct: err=%s\n", err)
		return c.Error(http.StatusInternalServerError, err)
	}
	defer request.Body.Close()
	event, err := github.ParseWebHook(github.WebHookType(request), payload)
	if err != nil {
		log.Printf("could not parse webhook: err=%s\n", err)
		return c.Error(http.StatusInternalServerError, err)
	}
	repoName := ""
	var myEvent eventgrid.Event
	myDate := date.Time{Time: time.Now()}
	var events []eventgrid.Event
	var myClient = eventgrid.New()
	switch e := event.(type) {
	case *github.PullRequestEvent:
		if e.Action != nil {
			repoName = *e.Repo.FullName
			fmt.Printf("Repository Name: %s", *e.Repo.FullName)
		}
	case *github.LabelEvent:
		if e.Action != nil {
			myEvent = eventgrid.Event{
				EventType:       to.StringPtr(os.Getenv("EVENT_TYPE")),
				EventTime:       &myDate,
				ID:              to.StringPtr(os.Getenv("ID")),
				Data:            e,
				Subject:         e.Label.URL,
				DataVersion:     to.StringPtr(""),
				MetadataVersion: to.StringPtr("1"),
			}
			events = append(events, myEvent)
			repoName = *e.Repo.FullName
			result, err := eventgrid.BaseClient.PublishEvents(myClient, request.Context(), "/specsla.westus2-1.eventgrid.azure.net/api/events", events)
			if err != nil {
				log.Printf("could not parse webhook: err=%s\n", err)
				return c.Error(result.Response.StatusCode, err)
			}
		}
	default:
		log.Printf("unknown event type %s\n", github.WebHookType(request))
		//return err
	}
	return c.Render(200, render.JSON(map[string]string{"message": "Welcome to Buffalo!", "repo name": repoName}))

}
