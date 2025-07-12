package framework

import (
	"flag"
)

var (
	// preserveOnFailure controls whether test clusters should be preserved when tests fail
	preserveOnFailure = flag.Bool("preserve-on-failure", false, "Preserve test clusters and resources when tests fail for debugging purposes")
)

// ShouldPreserveOnFailure returns true if test clusters should be preserved when tests fail
func ShouldPreserveOnFailure() bool {
	return *preserveOnFailure
}
