package mw

import (
	"fmt"
	"net/http"
	"testing"
)

type mockLogger string

func (l *mockLogger) Println(v ...interface{}) {
	*l = mockLogger(fmt.Sprint(v...))
}

func Test_PanicRecover(t *testing.T) {
	t.Run("panic recover middleware should be able to catch a panic in the next handlers and report to the log", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Error("the code should not panic because it is wrapped with PanicRecover middleware")
			}
		}()
		logger := new(mockLogger)
		handler := PanicRecover(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("it should panic") }))
		handler.ServeHTTP(nil, nil)
		if *logger != "Recovered from panic:it should panic" {
			t.Errorf("should be recovered from panic and report an error to the log with correct message")
		}
	})
}
