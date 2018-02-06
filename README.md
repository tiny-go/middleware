# go-middleware

[![Build Status][circleci-badge]][circleci-link]
[![Report Card][report-badge]][report-link]
[[!GoCover][cover-badge]][cover-link]

Golang HTTP middleware

### Examples

Build the handler using middleware chaining functions:
- `New()` - start the chain. Can accept 0 (zero) or more arguments.
- `Use()` - add middleware to existing chain.
- `Then()` - set the final handler (which is `http.Handler`).

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
			New(mw.PanicRecover(log.New(os.Stdout, "", 0))).
			Use(mw.BodyClose).
			Then(final),
	)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

```

[circleci-badge]: https://circleci.com/gh/Alma-media/go-middleware.svg?style=shield
[circleci-link]: https://circleci.com/gh/Alma-media/go-middleware
[report-badge]: https://goreportcard.com/badge/github.com/Alma-media/go-middleware
[report-link]: https://goreportcard.com/report/github.com/Alma-media/go-middleware
[cover-badge]: https://gocover.io/_badge/github.com/Alma-media/go-middleware
[cover-link]: https://gocover.io/github.com/Alma-media/go-middleware
