package dreck

import (
	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"

	"github.com/google/go-github/github"
)

// findReviewers will retrieve the files in files with the function f and returns possible reviewers in the map.
func findReviewers(files []*github.CommitFile, owners string, f func(path string) ([]byte, error)) map[string]string {

	victims := make(map[string]string)

File:
	for _, fi := range files {
		paths := ownersPaths(*fi.Filename, owners)
		// Find nearest OWNERS files.
		for _, p := range paths {
			log.Infof("Looking for OWNERS in %s", p)
			buf, err := f(p)
			if err != nil {
				continue
			}

			var config types.DreckConfig
			if err := parseConfig(buf, &config); err != nil {
				continue
			}
			for _, r := range config.Reviewers {
				victims[r] = p
			}
			continue File
		}
	}
	return victims
}
