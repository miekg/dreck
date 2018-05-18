package dreck

import (
	"path"
)

// ownersPaths returns all directories included in p, with owners suffixed, this may
// return duplicate paths.
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

/*
/plugin/auto/a/d.go
/plugin/auto/a/b.go
/plugin/auto/c.go
/plugin/trace/b.go
/plugin/README.md

Get each dirname and count the number of time you see it.

O(n2), but the amount is small.

/plugin/auto/a/d.go
/plugin/auto/a/b.go
/plugin/auto/c.go
/plugin/trace/b.go
README.md

full match on dirnames; dirname everything:

dirname first: /plugin/auto/a/d.go -> /plugin/auto/a
check: 2 in list

next: /plugin/auto/a/b.go -> plugin/auto/a
check: 2 in list (can skip because saving in map)

next: /plugin/auto/c.go -> plugin/auto
check: 1 in list

next: /plugin/trace/b.go -> plugin/trace
check: 1 in list

next: /plugin/README.md -> plugin
check: 1 in list

plugin/auto/a: 2
plugin/auto: 1
plugin: 1

sort on highest number: start lookin for owners there.
*/

// mostSpecific will get a tally of each path on how aften that specific one
// is contained in the entire set of paths.
func mostSpecific(p []string) map[string]int {
	m := make(map[string]int)

	for i := 0; i < len(p); i++ {
		if _, ok := m[p[i]]; ok {
			// already seen this path.
			continue
		}
		for j := 0; j < len(p); j++ {
			if p[i] == p[j] {
				m[p[i]]++
			}
		}
	}

	return m
}

// sortOnOccurence sorts the map[string]int on the integers and returns a [][]string that is
// indexed on the number of occurences and contains the paths that have that many occurences.
// The first (zero-th) element is always empty.
func sortOnOccurence(m map[string]int) [][]string {
	ret := make([][]string, len(m))

	for p, v := range m {
		ret[v] = append(ret[v], p)
	}
	return ret
}
