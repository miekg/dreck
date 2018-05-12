package dreck

import (
	"fmt"
	"regexp"
	"strings"
)

// Rule write an alias' from to to.
type Rule struct {
	from    *regexp.Regexp
	replace string
}

// Expand will use R to expand the command.
func (r Rule) Expand(src string) string {
	return r.from.ReplaceAllString(src, r.replace)
}

// NewAlias inspects command to see if it is a correct alias.
func NewAlias(command string) (Rule, error) {
	var err error
	splits := strings.Split(command, sep)
	if len(splits) != 2 {
		return Rule{}, fmt.Errorf("could not find alias in %s", command)
	}
	r := Rule{replace: splits[1]}
	r.from, err = regexp.Compile(splits[0])
	if err != nil {
		return r, err
	}

	return r, nil
}

const sep = " -> "
