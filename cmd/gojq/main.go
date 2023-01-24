// gojq - Go implementation of jq
package main

import (
	"os"

	"github.com/gozelle/jq/cli"
)

func main() {
	os.Exit(cli.Run())
}
