package common

import (
	"context"
	"encoding/json"
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"os"
)

func ParseGithubEvent(event interface{}) {
	body, err := ioutil.ReadFile(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil {
		panic(err)
	}

	//var event github.IssueEvent
	err = json.Unmarshal(body, event)
	if err != nil {
		panic(err)
	}
}

func MakeClients(ctx context.Context, token string) (clientWithToken, client *github.Client) {
	//token := os.Getenv("GITHUB_ACCESS_TOKEN")
	if token == "" {
		panic("empty github token")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts) // Transport
	clientWithToken = github.NewClient(tc)

	client = github.NewClient(nil) // 规避 API Rate limit
	return
}
