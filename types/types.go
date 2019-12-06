package types

type Repository struct {
	Owner Owner  `json:"owner"`
	Name  string `json:"name"`
}

type Owner struct {
	Login string `json:"login"`
	Type  string `json:"type"`
}

type InstallationRequest struct {
	Installation ID `json:"installation"`
}

type ID struct {
	ID int `json:"id"`
}

type IssueCommentOuter struct {
	Repository  Repository  `json:"repository"`
	Comment     Comment     `json:"comment"`
	Action      string      `json:"action"`
	Issue       Issue       `json:"issue,omitempty"`
	PullRequest PullRequest `json:"pull_request,omitempty"`
	InstallationRequest
}

type IssueLabel struct {
	Name string `json:"name"`
}

type Issue struct {
	Labels []IssueLabel `json:"labels"`
	Number int          `json:"number"`
	Title  string       `json:"title"`
	Locked bool         `json:"locked"`
	State  string       `json:"state"`
	Body   string       `json:"body,omitempty""`
	User   struct {
		Login string `json:"login"`
	}
}

type PullRequest struct {
	Body string `json:"body,omitempty""`
	User struct {
		Login string `json:"login"`
	}
	Number int `json:"number"`
}

type Comment struct {
	Body     string `json:"body"`
	IssueURL string `json:"issue_url"`
	User     struct {
		Login string `json:"login"`
	}
}

type Action struct {
	Type  string
	Value string
}

// DreckConfig holds the configuration from the .dreck.yaml and CODEOWNERS file.
type DreckConfig struct {
	CodeOwners []string
	Aliases    []string
	Features   []string
}
