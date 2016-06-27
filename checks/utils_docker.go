package checks

import (
	"errors"
	"strings"
)

func getContainerName(names []string) (string, error) {

	// remove prefix '/'
	for _, name := range names {
		namePrefixRemoved := name[1:]

		// find container without '/' within name
		if len(strings.Split(namePrefixRemoved, "/")) == 1 {
			return namePrefixRemoved, nil
		}
	}

	return "", errors.New("check utils: unable to find container name")
}
