// generated by dispel v1; DO NOT EDIT

package admin

import (
	"errors"
	"net/http"
)

// HandlerRegisterer is the interface implemented by objects that can register a http handler
// for an http route.
type HandlerRegisterer interface {
	RegisterHandler(routeName string, handler http.Handler)
}

// registerHandlerFunc is an adapter to use funcs as HandlerRegisterer.
type registerHandlerFunc func(routeName string, handler http.Handler)

// RegisterHandler calls f(routeName, handler).
func (f registerHandlerFunc) RegisterHandler(routeName string, handler http.Handler) {
	f(routeName, handler)
}

// RouteParamGetter is the interface implemented by objects that can retrieve
// the value of a parameter of a route, by name.
type RouteParamGetter interface {
	GetRouteParam(r *http.Request, name string) string
}

// HTTPEncoder is the interface implemented by objects that can encode values to a http response,
// with the specified http status.
//
// Implementors must handle nil data.
type HTTPEncoder interface {
	Encode(w http.ResponseWriter, r *http.Request, data interface{}, code int) error
}

// HTTPDecoder is the interface implemented by objects that can decode data received from a http request.
//
// Implementors have to close the request.Body.
// Decode() shouldn't write to http.ResponseWriter: it's up to the caller to e.g, handle errors.
type HTTPDecoder interface {
	Decode(http.ResponseWriter, *http.Request, interface{}) error
}

// errorHTTPHandlerFunc defines the signature of the generated http handlers used in registerHandlers().
//
// The basic contract of this handler is it write the status code to w (and the body, if any), unless an error is returned;
// in this case, the caller has to write to w.
type errorHTTPHandlerFunc func(w http.ResponseWriter, r *http.Request) (status int, err error)

// registerHandlers registers resource handlers for each unique named route.
// registerHandlers must be called after the registerRoutes().
func registerHandlers(hr HandlerRegisterer, rpg RouteParamGetter, sph *ServerPoolHandler, hd HTTPDecoder, he HTTPEncoder, ehhf func(errorHTTPHandlerFunc) http.Handler) {
	hr.RegisterHandler(routeFileservers, &MethodHandler{
		Get: ehhf(func(w http.ResponseWriter, r *http.Request) (int, error) {
			status, vresp, err := sph.getFileservers(w, r)
			if err != nil {
				return status, err
			}
			return status, he.Encode(w, r, vresp, status)
		}),
	})
	hr.RegisterHandler(routeServers, &MethodHandler{
		Get: ehhf(func(w http.ResponseWriter, r *http.Request) (int, error) {
			status, vresp, err := sph.getServers(w, r)
			if err != nil {
				return status, err
			}
			return status, he.Encode(w, r, vresp, status)
		}),
		Post: ehhf(func(w http.ResponseWriter, r *http.Request) (int, error) {
			var vreq CreateRandomServerIn
			if err := hd.Decode(w, r, &vreq); err != nil {
				return http.StatusBadRequest, err
			}
			status, vresp, err := sph.postServers(w, r, &vreq)
			if err != nil {
				return status, err
			}
			return status, he.Encode(w, r, vresp, status)
		}),
		Delete: ehhf(func(w http.ResponseWriter, r *http.Request) (int, error) {
			status, vresp, err := sph.deleteServers(w, r)
			if err != nil {
				return status, err
			}
			return status, he.Encode(w, r, vresp, status)
		}),
	})
	hr.RegisterHandler(routeServersOne, &MethodHandler{
		Get: ehhf(func(w http.ResponseWriter, r *http.Request) (int, error) {
			serverPort := rpg.GetRouteParam(r, "server-port")
			if serverPort == "" {
				return http.StatusBadRequest, errors.New("empty route parameter \"server-port\"")
			}
			status, vresp, err := sph.getServersOne(w, r, serverPort)
			if err != nil {
				return status, err
			}
			return status, he.Encode(w, r, vresp, status)
		}),
		Put: ehhf(func(w http.ResponseWriter, r *http.Request) (int, error) {
			serverPort := rpg.GetRouteParam(r, "server-port")
			if serverPort == "" {
				return http.StatusBadRequest, errors.New("empty route parameter \"server-port\"")
			}
			var vreq CreateServerIn
			if err := hd.Decode(w, r, &vreq); err != nil {
				return http.StatusBadRequest, err
			}
			status, vresp, err := sph.putServersOne(w, r, serverPort, &vreq)
			if err != nil {
				return status, err
			}
			return status, he.Encode(w, r, vresp, status)
		}),
		Delete: ehhf(func(w http.ResponseWriter, r *http.Request) (int, error) {
			serverPort := rpg.GetRouteParam(r, "server-port")
			if serverPort == "" {
				return http.StatusBadRequest, errors.New("empty route parameter \"server-port\"")
			}
			status, vresp, err := sph.deleteServersOne(w, r, serverPort)
			if err != nil {
				return status, err
			}
			return status, he.Encode(w, r, vresp, status)
		}),
	})
	hr.RegisterHandler(routeServersOneMounts, &MethodHandler{
		Get: ehhf(func(w http.ResponseWriter, r *http.Request) (int, error) {
			serverPort := rpg.GetRouteParam(r, "server-port")
			if serverPort == "" {
				return http.StatusBadRequest, errors.New("empty route parameter \"server-port\"")
			}
			status, vresp, err := sph.getServersOneMounts(w, r, serverPort)
			if err != nil {
				return status, err
			}
			return status, he.Encode(w, r, vresp, status)
		}),
		Post: ehhf(func(w http.ResponseWriter, r *http.Request) (int, error) {
			serverPort := rpg.GetRouteParam(r, "server-port")
			if serverPort == "" {
				return http.StatusBadRequest, errors.New("empty route parameter \"server-port\"")
			}
			var vreq CreateMountIn
			if err := hd.Decode(w, r, &vreq); err != nil {
				return http.StatusBadRequest, err
			}
			status, vresp, err := sph.postServersOneMounts(w, r, serverPort, &vreq)
			if err != nil {
				return status, err
			}
			return status, he.Encode(w, r, vresp, status)
		}),
		Delete: ehhf(func(w http.ResponseWriter, r *http.Request) (int, error) {
			serverPort := rpg.GetRouteParam(r, "server-port")
			if serverPort == "" {
				return http.StatusBadRequest, errors.New("empty route parameter \"server-port\"")
			}
			status, vresp, err := sph.deleteServersOneMounts(w, r, serverPort)
			if err != nil {
				return status, err
			}
			return status, he.Encode(w, r, vresp, status)
		}),
	})
	hr.RegisterHandler(routeServersOneMountsOne, &MethodHandler{
		Get: ehhf(func(w http.ResponseWriter, r *http.Request) (int, error) {
			serverPort := rpg.GetRouteParam(r, "server-port")
			if serverPort == "" {
				return http.StatusBadRequest, errors.New("empty route parameter \"server-port\"")
			}
			mountId := rpg.GetRouteParam(r, "mount-id")
			if mountId == "" {
				return http.StatusBadRequest, errors.New("empty route parameter \"mount-id\"")
			}
			status, vresp, err := sph.getServersOneMountsOne(w, r, serverPort, mountId)
			if err != nil {
				return status, err
			}
			return status, he.Encode(w, r, vresp, status)
		}),
		Delete: ehhf(func(w http.ResponseWriter, r *http.Request) (int, error) {
			serverPort := rpg.GetRouteParam(r, "server-port")
			if serverPort == "" {
				return http.StatusBadRequest, errors.New("empty route parameter \"server-port\"")
			}
			mountId := rpg.GetRouteParam(r, "mount-id")
			if mountId == "" {
				return http.StatusBadRequest, errors.New("empty route parameter \"mount-id\"")
			}
			status, vresp, err := sph.deleteServersOneMountsOne(w, r, serverPort, mountId)
			if err != nil {
				return status, err
			}
			return status, he.Encode(w, r, vresp, status)
		}),
	})
}
