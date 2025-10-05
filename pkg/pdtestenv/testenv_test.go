package pdtestenv

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	Teardown()
	os.Exit(code)
}
