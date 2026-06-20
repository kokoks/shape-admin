package main
import (
	"os"
	"testing"
)

func TestEnvLoading(t *testing.T) {
	os.Setenv("TEST_VAR", "test")
	val := os.Getenv("TEST_VAR")
	if val != "test" {
		t.Error("Expected test value")
	}
}