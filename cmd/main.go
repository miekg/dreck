// This is a command that just prints the version of dreck.
// We use this to test the asset and docker image creation.
package main

import (
	"fmt"

	"github.com/miekg/dreck"
)

func main() {
	fmt.Printf("dreck-go: %s", dreck.Version)
}
