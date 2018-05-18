package dreck

import (
	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"

	"github.com/google/go-github/github"
)

func (d Dreck) findReviewers(files []*github.CommitFile, puller string, f func(path string) ([]byte, error)) (string, string) {
	allFiles := []string{}
	for _, fi := range files {
		paths := ownersPaths(*fi.Filename, d.owners)
		allFiles = append(allFiles, paths...)
	}
	specific := mostSpecific(allFiles)
	order := sortOnOccurence(specific)

	log.Infof("Looking for reviewers excluding %s", puller)

	// order now contains the best owners file paths (OWNER files may not exist) to select
	// the owners from, so we go from heighest to lowest to select an owner.

	for i := range order {
		files := order[i]
		for j := range files {
			log.Infof("Looking for OWNERS in %s", files[j])
			buf, err := f(files[j])
			if err != nil {
				continue
			}

			var config types.DreckConfig
			if err := parseConfig(buf, &config); err != nil {
				continue
			}
			// Valid OWNERS file, we will return the first non-PR person we find.
			for _, r := range config.Reviewers {
				if r != puller {
					return r, files[j]
				}
			}
		}
	}
	return "", ""
}
