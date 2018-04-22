package mw

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

const (
	contentTypeHeader = "Content-Type"
	acceptHeader      = "Accept"
)

// codecKey is a private unique key that is used to put/get codec from the context.
type codecKey struct{ kind string }

// Encoder is responsible for data encoding.
type Encoder interface {
	Encode(v interface{}) error
}

// Decoder is responsible for data decoding.
type Decoder interface {
	Decode(v interface{}) error
}

// Codec can create Encoder(s) and Decoder(s) with provided io.Reader/io.Writer.
type Codec interface {
	// Ecoder instantiates the ecnoder part of the codec with provided writer.
	Encoder(w io.Writer) Encoder
	// Decoder instantiates the decoder part of the codec with provided reader.
	Decoder(r io.Reader) Decoder
	// MimeType returns the (main) mime type of the codec.
	MimeType() string
}

// Codecs represents any kind of codec registry, thus it can be global registry
// or a custom list of codecs that is supposed to be used for particular route.
type Codecs interface {
	// Lookup should find appropriate Codec by MimeType or return nil if not found.
	Lookup(mimeType string) Codec
}

// CodecFromList middleware searches for suitable request/response codecs according to
// "Content-Type" and "Accept" headers and puts the correct codecs into the context.
// NOTE: do not use current function without PanicRecover middleware.
func CodecFromList(codecs Codecs) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqCodec, resCodec Codec
			// get request codec
			if reqCodec = codecs.Lookup(r.Header.Get(contentTypeHeader)); reqCodec == nil {
				panic(
					NewStatusError(
						http.StatusBadRequest,
						fmt.Sprintf("unsupported request codec: %q", r.Header.Get(contentTypeHeader)),
					),
				)
			}
			r = r.WithContext(context.WithValue(r.Context(), codecKey{"req"}, reqCodec))
			// get response codec
			if resCodec = codecs.Lookup(r.Header.Get(acceptHeader)); resCodec == nil {
				panic(
					NewStatusError(
						http.StatusBadRequest,
						fmt.Sprintf("unsupported response codec: %q", r.Header.Get(acceptHeader)),
					),
				)
			}
			r = r.WithContext(context.WithValue(r.Context(), codecKey{"res"}, resCodec))
			// call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequestCodecFromContext pulls the Codec from a request context or or returns nil.
func RequestCodecFromContext(ctx context.Context) Codec {
	codec, _ := ctx.Value(codecKey{"req"}).(Codec)
	return codec
}

// ResponseCodecFromContext pulls the Codec from a request context or returns nil.
func ResponseCodecFromContext(ctx context.Context) Codec {
	codec, _ := ctx.Value(codecKey{"res"}).(Codec)
	return codec
}
