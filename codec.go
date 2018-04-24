package mw

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tiny-go/codec"
)

const (
	contentTypeHeader = "Content-Type"
	acceptHeader      = "Accept"
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
// "Content-Type" and "Accept" headers and puts the correct codecs into the context.
// NOTE: do not use current function without PanicRecover middleware.
func Codec(codecs Codecs) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqCodec, resCodec codec.Codec
			// get request codec
			if reqCodec = codecs.Lookup(r.Header.Get(contentTypeHeader)); reqCodec == nil {
				panic(NewStatusError(
					http.StatusBadRequest,
					fmt.Errorf("unsupported request codec: %q", r.Header.Get(contentTypeHeader)),
				))
			}
			r = r.WithContext(context.WithValue(r.Context(), codecKey{"req"}, reqCodec))
			// get response codec
			if resCodec = codecs.Lookup(r.Header.Get(acceptHeader)); resCodec == nil {
				panic(NewStatusError(
					http.StatusBadRequest,
					fmt.Errorf("unsupported response codec: %q", r.Header.Get(acceptHeader)),
				))
			}
			r = r.WithContext(context.WithValue(r.Context(), codecKey{"res"}, resCodec))
			// call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequestCodecFromContext pulls the Codec from a request context or or returns nil.
func RequestCodecFromContext(ctx context.Context) codec.Codec {
	codec, _ := ctx.Value(codecKey{"req"}).(codec.Codec)
	return codec
}

// ResponseCodecFromContext pulls the Codec from a request context or returns nil.
func ResponseCodecFromContext(ctx context.Context) codec.Codec {
	codec, _ := ctx.Value(codecKey{"res"}).(codec.Codec)
	return codec
}
