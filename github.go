package dreck

import (
	"context"

	"github.com/miekg/dreck/auth"

	"github.com/google/go-github/github"
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
func (d Dreck) githubFile(owner, repository string) ([]byte, error)
	maintainersFile := fmt.Sprintf("https://github.com/%s/%s/raw/master/%s", owner, repository, d.owners)

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	req, _ := http.NewRequest(http.MethodGet, maintainersFile, nil)

	res, resErr := client.Do(req)
	if resErr != nil {
		return nil, resErr
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Status code: %d while fetching maintainers list (%s)", res.StatusCode, maintainersFile)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, _ := ioutil.ReadAll(res.Body)
