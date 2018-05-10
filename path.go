package dreck

import (
	"path"
)

// ownersPaths returns all directories included in p, with owners suffixed.
func ownersPaths(p, owner string) []string {
	s := []string{}
	p1 := p
	for {
		p1 = path.Dir(p1)
		if p1 == "." || p1 == "/" {
			s = append(s, path.Join("/", owner))
			return s
		}
		s = append(s, path.Join(p1, owner))
	}
}
