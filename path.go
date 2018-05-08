package dreck

import "path"

// ownersPaths returns all directories includes in p, with p.owners suffixed.
func ownersPaths(p, owner string) []string {
	s := []string{owner}
	p1 := p
	for {
		p1 = path.Dir(p1)
		if p1 == "." || p1 == "/" {
			return s
		}
		s = append(s, path.Join(p1, owner))
	}
	return nil
}
