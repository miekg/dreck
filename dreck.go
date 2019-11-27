package dreck

import (
	"bufio"
	"bytes"
	"strings"

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

	if err := yaml.Unmarshal(buf, &config); err != nil {
		return nil, err
	}

	// grap toplevel CODEOWNERS file and parse that
	buf, err = githubFile(owner, repository, "CODEOWNERS")
	if err != nil {
		return nil, err
	}
	config.CodeOwners, err = parseOwners(buf)
	return &config, err
}

func parseOwners(buf []byte) ([]string, error) {
	// simple line, by line based format
	//
	// # In this example, @doctocat owns any files in the build/logs
	// # directory at the root of the repository and any of its
	// # subdirectories.
	// /build/logs/ @doctocat

	scanner := bufio.NewScanner(bytes.NewReader(buf))
	users := map[string]struct{}{}
	for scanner.Scan() {
		text := scanner.Text()
		if len(text) == 0 {
			continue
		}
		if text[0] == '#' {
			continue
		}
		ele := strings.Fields(text)
		if len(ele) == 0 {
			continue
		}

		// ok ele[0] is the path, the rest are (in our case) github usernames prefixed with @
		for _, s := range ele[1:] {
			if len(s) <= 1 {
				continue
			}
			users[s[1:]] = struct{}{}
		}

	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	u := []string{}
	for k, _ := range users {
		u = append(u, k)
	}
	return u, nil

}

const (
	Aliases = "aliases" // Aliases enables alias expansion.
	Exec    = "exec"    // Exec enables the exec command.
)

// Trigger is the prefix that triggers action from this bot.
const Trigger = "/"
