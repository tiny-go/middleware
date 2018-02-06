# go-middleware

[![Build Status][circleci-badge]][circleci-link]
[![Report Card][report-badge]][report-link]


Golang HTTP middleware

### Examples

- using middleware chaining functions:

```go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Alma-media/go-middleware"
)

func main() {
	var final http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	}

	http.Handle(
		"/",
		mw.
      // use New() to start the chain
			New(mw.PanicRecover(log.New(os.Stdout, "", 0))).
      // Use() - to add middleware
			Use(mw.BodyClose).
      // Then() - to set the final handler
			Then(final),
	)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

```

[circleci-badge]: https://circleci.com/gh/Alma-media/go-middleware.svg?style=shield
[circleci-link]: https://circleci.com/gh/Alma-media/go-middleware
[report-badge]: https://goreportcard.com/badge/github.com/Alma-media/go-middleware
[report-link]: https://goreportcard.com/report/github.com/Alma-media/go-middleware
