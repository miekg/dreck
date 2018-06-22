package dreck

import (
	"os/user"
	"strconv"
)

// userID looks up the uid and gid for the user named name.
func userID(name string) (uint32, uint32, error) {
	u, err := user.Lookup(name)
	if err != nil {
		return 65534, 65534, err
	}

	return toUint32(u.Uid), toUint32(u.Gid), nil
}

func toUint32(s string) uint32 {
	v, _ := strconv.Atoi(s)
	return uint32(v)
}
