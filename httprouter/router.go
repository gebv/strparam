package httprouter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/gebv/strparam"
)

// NewRouter returns routing.
func NewRouter() *Router {
	return &Router{
		store:       strparam.NewStore(),
		handlersMap: make(map[string]http.HandlerFunc),
	}
}

// Router implements http routing.
type Router struct {
	store           *strparam.Store
	handlersMap     map[string]http.HandlerFunc
	ErrorHandler    http.HandlerFunc
	NotFoundHandelr http.HandlerFunc
}

// Add adds a handle for the specified path and method.
func (r *Router) Add(method, addPath string, h http.HandlerFunc) error {
	// formatting the input value
	method = strings.ToUpper(method)

	// have come to believe that we get the expected value
	if !allowedMethods[method] {
		// NOTE: you can extend the supported methods
		return fmt.Errorf("does not support method %q", method)
	}

	// TODO: need right validator
	// some basic checks
	if len(addPath) == 0 {
		return errors.New("addPath cannot be empty")
	}
	if !strings.HasPrefix(addPath, "/") {
		return errors.New("addPath must start with a slash")
	}
	if strings.HasPrefix(addPath, "/..") || strings.HasPrefix(addPath, "./.") {
		// TODO: more detail or find a ready solution
		// https://tools.ietf.org/html/rfc3986
		return errors.New("path cannot be has './.' or '/..'")
	}

	// forming an internal key
	routeKey := ":" + method + ":" + addPath
	routePattern, err := strparam.Parse(routeKey)
	if err != nil {
		return errors.Wrap(err, "failed parse route key")
	}

	xRoutePattern := &strparam.Pattern{
		Tokens:    strparam.Tokens{},
		NumParams: routePattern.NumParams,
	}

	for _, token := range routePattern.Tokens {
		if token.Mode == strparam.CONST {
			fields := strings.Split(token.Raw, "/")
			for i, field := range fields {
				if field != "" {
					xRoutePattern.Tokens = append(xRoutePattern.Tokens, strparam.ConstToken(field))
				}
				if i < len(fields)-1 {
					xRoutePattern.Tokens = append(xRoutePattern.Tokens, strparam.SeparatorToken("/"))
				}
			}
		} else {
			xRoutePattern.Tokens = append(xRoutePattern.Tokens, token)
		}
	}

	r.store.AddPattern(xRoutePattern)

	// save the handler by hash of pattern
	// if exists returns error
	routePatternID := strparam.ListTokensSchemaString(xRoutePattern.Tokens)
	if _, exists := r.handlersMap[routePatternID]; exists {
		return fmt.Errorf("route %q already exists", addPath)
	}
	r.handlersMap[routePatternID] = h

	return nil
}

// Find returns handler and parsed params if found patter from the input request path and method.
// NOTE: case sensitive (it is correct for http routing in the general case?)
func (r *Router) Find(method, requestPath string) (http.HandlerFunc, map[string]string, error) {
	// formatting the input value
	method = strings.ToUpper(method)

	// have come to believe that we get the expected value
	if !allowedMethods[method] {
		// NOTE: you can extend the supported methods
		return nil, nil, fmt.Errorf("does not support method %q", method)
	}

	// forming an internal key
	routeKey := ":" + method + ":" + requestPath

	// looking for a matche pattern
	// returns error if not exists
	foundPattern := r.store.Find(routeKey)
	if foundPattern == nil {
		return nil, nil, errors.New("no matching patterns")
	}

	// parse the input string by pattern
	matched, paramsList := foundPattern.Lookup(routeKey)
	if !matched {
		// TODO: when can this happen?
		return nil, nil, errors.New("no matching patterns")
	}

	routePatternID := strparam.ListTokensSchemaString(foundPattern.Tokens)

	routeHandler, found := r.handlersMap[routePatternID]
	if !found {
		return nil, nil, errors.New("not found handler by matched route")
	}

	// transform the parameters into a map
	paramsMap := make(map[string]string, len(paramsList))
	for _, item := range paramsList {
		paramsMap[item.Name] = item.Value
	}

	return routeHandler, paramsMap, nil
}

// HandleFunc processing of request.
func (r *Router) HandleFunc(w http.ResponseWriter, req *http.Request) {
	h, params, err := r.Find(req.Method, req.URL.Path)
	if err != nil {
		if err.Error() == "no matching patterns" {
			r.NotFoundHandelr(w, req)
			return
		}
		r.ErrorHandler(w, req)
		return
	}
	req = req.WithContext(setParsedParamsToCtx(req.Context(), params))
	h(w, req)
}

// ServeHTTP implements http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.HandleFunc(w, req)
}

var _ http.Handler = (*Router)(nil)

var allowedMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodPost:    true,
	http.MethodPut:     true,
	http.MethodHead:    true,
	http.MethodDelete:  true,
	http.MethodOptions: true,
}
