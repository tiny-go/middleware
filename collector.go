package mw

// Collector should contain a list of Middleware funcs that are supposed to be
// applied to the request with given HTTP method.
type Collector interface {
	// AddMiddleware should make the provided chain available by HTTP method.
	AddMiddleware(method string, chain ...Middleware) Collector
	// Middleware should return the list of middleware functions registered for provided method.
	Middleware(method string) []Middleware
}

// List provides ability to register/unregister middleware for HTTP methods.
type List struct {
	middleware map[string][]Middleware
}

// NewList is a constructor func for middleware List.
func NewList() *List {
	return &List{
		middleware: make(map[string][]Middleware),
	}
}

// AddMiddleware adds middleware funcs to existing ones for provided HTTP method.
func (l *List) AddMiddleware(method string, chain ...Middleware) Collector {
	if _, ok := l.middleware[method]; !ok {
		l.middleware[method] = []Middleware{}
	}
	l.middleware[method] = append(l.middleware[method], chain...)
	// return itself in order to use func in a chain
	return l
}

// Middleware returns middleware func registered for provided method or an empty list.
func (l *List) Middleware(method string) []Middleware {
	if mw, ok := l.middleware[method]; ok {
		return mw
	}
	return []Middleware{}
}
