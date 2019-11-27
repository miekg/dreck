package dreck

import (
	"github.com/miekg/dreck/types"

	"github.com/caddyserver/caddy/caddyhttp/httpserver"
	yaml "gopkg.in/yaml.v2"
)

// Dreck is a plugin that handles Github Issues and Pull Requests for you.
type Dreck struct {
	Next httpserver.Handler

	clientID string
	key      string

	owners   string
	secret   string
	path     string            // when should dreck trigger, default to '/dreck'
	hmac     bool              // validate HMAC on the webhook
	strategy string            // how to merge when we merge
	user     string            // user to use to exec commands
	env      map[string]string // environment to give to commands
}

// New returns a new, initialized Dreck.
func New() Dreck {
	d := Dreck{}
	d.owners = ".dreck.yaml"
	d.path = "/dreck"
	d.strategy = mergeSquash
	d.env = make(map[string]string)

	return d
}

func (d Dreck) getConfig(owner string, repository string) (*types.DreckConfig, error) {
	var config types.DreckConfig

	buf, err := githubFile(owner, repository, d.owners)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(bytesOut, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

const (
	Aliases = "aliases" // Aliases enables alias expansion.
	Exec    = "exec"    // Exec enables the exec command.
)

// Trigger is the prefix that triggers action from this bot.
const Trigger = "/"
