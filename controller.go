package mw

// Controller represents simple HTTP controller containing middleware for each method.
type Controller interface {
	// AddMiddleware should make the provided chain available by HTTP method.
	AddMiddleware(method string, chain ...Middleware) Controller
	// Middleware should return the list of middleware functions registered for provided method.
	Middleware(method string) Middleware
}

// BaseController provides ability to register/unregister middleware for HTTP methods.
//This is a basic Controller implementation.
type BaseController struct {
	middleware map[string]Middleware
}

// NewBaseController is a constructor func for a basic HTTP controller.
func NewBaseController() *BaseController {
	return &BaseController{
		middleware: make(map[string]Middleware),
	}
}

// AddMiddleware adds middleware funcs to existing ones for provided HTTP method.
func (bc *BaseController) AddMiddleware(method string, chain ...Middleware) Controller {
	if _, ok := bc.middleware[method]; !ok {
		// create new middleware
		bc.middleware[method] = New(chain...)
	} else {
		// upgrade existing middleware and replace the old one
		bc.middleware[method] = bc.middleware[method].Use(chain...)
	}
	// return itself in order to use func in a chain
	return bc
}

// Middleware returns middleware func registered for provided method or an empty list.
func (bc *BaseController) Middleware(method string) Middleware {
	if mw, ok := bc.middleware[method]; ok {
		return mw
	}
	return New()
}

// Init does nothing. This is a default function to avoid explicit declaration
// when controller does not require any Init logic.
func (bc *BaseController) Init() error { return nil }
