package dreck

import (
	"github.com/google/go-github/github"

	"github.com/miekg/dreck/log"
)

// logRateLimit spews out a log line about the current rate limit rate.
func logRateLimit(resp *github.Response) error {

	log.Warningf("Rate limiting: %s", resp.Rate)
	return nil
}
