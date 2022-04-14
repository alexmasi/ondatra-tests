package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/openconfig/ondatra"
	kinit "github.com/openconfig/ondatra/knebind/init"
)

// TestMain is the first thing that's executed upon running `go test ...`
func TestMain(m *testing.M) {
	fmt.Println(os.Args)
	ondatra.RunTests(m, kinit.Init)
}
