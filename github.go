package dreck

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/miekg/dreck/auth"

	"github.com/google/go-github/v28/github"
)

func githubIssueComment(body string) *github.IssueComment {
	return &github.IssueComment{
		Body: &body,
	}
}

func (d Dreck) newClient(installation int) (*github.Client, context.Context, error) {
	ctx := context.Background()
	token, err := auth.MakeAccessTokenForInstallation(d.clientID, d.key, installation)
	if err != nil {
		return nil, ctx, err
	}

	client := auth.MakeClient(ctx, token)
	return client, ctx, nil
}

// githubFile returns the file from github or an error if nothing is found.
func githubFile(owner, repository, path string) ([]byte, error) {
	file := fmt.Sprintf("https://github.com/%s/%s/raw/master/%s", owner, repository, path)
	client := http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, file, nil)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %d while fetching maintainers list (%s)", res.StatusCode, file)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	return ioutil.ReadAll(res.Body)
}
