package mw

import (
	"net/http"
	"testing"
)

var _ Error = StatusError{}

func TestStatusError(t *testing.T) {
	err := NewStatusError(http.StatusNotFound, "not found")
	t.Run("test if StatusError returns expected error code", func(t *testing.T) {
		if err.Code() != http.StatusNotFound {
			t.Errorf("error code %d was expected to be %d", err.Code(), http.StatusNotFound)
		}
		if err.Error() != "not found" {
			t.Errorf("error message %q was expected to be %q", err.Error(), "not found")
		}
	})
}
