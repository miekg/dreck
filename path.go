package dreck

import (
	"path"
	"sort"
)

// ownersPaths returns all directories included in p, with owners suffixed, this may
// return duplicate paths.
func (d Dreck) ownersPaths(p string) []string {
	s := []string{}
	p1 := p
	for {
		p1 = path.Dir(p1)
		if p1 == "." || p1 == "/" {
			s = append(s, path.Join("/", d.owners))
			return s
		}
		s = append(s, path.Join(p1, d.owners))
	}
}

// mostSpecific will get a tally of each path on how often that specific one
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
// The first (zero-th) element is always empty and can contain or gaps.
func sortOnOccurence(m map[string]int) [][]string {
	// find the largest integer in the map, so we can size ret accordingly
	max := 0
	for _, v := range m {
		if v > max {
			max = v
		}
	}
	ret := make([][]string, max+1)

	for p, v := range m {
		ret[v] = append(ret[v], p)
	}
	// Sort longer paths first.
	for i := range ret {
		sort.Sort(ByLen(ret[i]))
	}

	return ret
}

type ByLen []string

func (a ByLen) Len() int           { return len(a) }
func (a ByLen) Less(i, j int) bool { return len(a[i]) > len(a[j]) }
func (a ByLen) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

//	sort.Sort(ByLen(s))
