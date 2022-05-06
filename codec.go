package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tiny-go/codec"
	"github.com/tiny-go/errors"
)

const (
	acceptHeader      = "Accept"
	contentTypeHeader = "Content-Type"
)

// codecKey is a private unique key that is used to put/get codec from the context.
type codecKey struct{ kind string }

// Codecs represents any kind of codec registry, thus it can be global registry
// or a custom list of codecs that is supposed to be used for particular route.
type Codecs interface {
	// Lookup should find appropriate Codec by MimeType or return nil if not found.
	Lookup(mimeType string) codec.Codec
}

// Codec middleware searches for suitable request/response codecs according to
// "Content-Type"/"Accept" headers and puts the correct codecs into the context.
func Codec(fn errors.HandlerFunc, codecs Codecs) Middleware {
	if fn == nil {
		fn = http.Error
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqCodec, resCodec codec.Codec
			// get request codec
			if reqCodec = codecs.Lookup(r.Header.Get(contentTypeHeader)); reqCodec == nil {
				fn(w, fmt.Sprintf("unsupported request codec: %q", r.Header.Get(contentTypeHeader)), http.StatusBadRequest)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), codecKey{"req"}, reqCodec))
			// get response codec
			if resCodec = codecs.Lookup(r.Header.Get(acceptHeader)); resCodec == nil {
				fn(w, fmt.Sprintf("unsupported response codec: %q", r.Header.Get(acceptHeader)), http.StatusBadRequest)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), codecKey{"res"}, resCodec))
			// call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequestCodecFromContext pulls the Codec from a request context or returns nil.
func RequestCodecFromContext(ctx context.Context) codec.Codec {
	codec, _ := ctx.Value(codecKey{"req"}).(codec.Codec)
	return codec
}

// ResponseCodecFromContext pulls the Codec from a request context or returns nil.
func ResponseCodecFromContext(ctx context.Context) codec.Codec {
	codec, _ := ctx.Value(codecKey{"res"}).(codec.Codec)
	return codec
}
