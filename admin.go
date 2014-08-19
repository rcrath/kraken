package kraken

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/vincent-petithory/kraken/fileserver"
)

type serverPoolAdminHandler struct {
	*ServerPool
	h      http.Handler
	router *mux.Router
}

const (
	routeServers                = "servers"
	routeServersSelf            = "servers.self"
	routeServersSelfAliases     = "servers.self.aliases"
	routeServersSelfAliasesSelf = "servers.self.aliases.self"
	routeFileservers            = "fileservers"
)

type AdminAPIErrorType string

const (
	apiErrTypeBadRequest  AdminAPIErrorType = "bad_request_error"
	apiErrTypeAPIInternal                   = "api_internal_error"
)

type AdminAPIError struct {
	Type AdminAPIErrorType `json:"type"`
	Msg  string            `json:"msg"`
}

func (e *AdminAPIError) String() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Msg)
}

func (e *AdminAPIError) Error() string {
	return e.Msg
}

func NewServerPoolAdminHandler(serverPool *ServerPool) *serverPoolAdminHandler {
	spah := serverPoolAdminHandler{ServerPool: serverPool}
	r := mux.NewRouter()
	apiRouter := r.PathPrefix("/api/").Subrouter()
	apiRouter.Handle("/servers", handlers.MethodHandler{
		"GET":    http.HandlerFunc(spah.getServers),
		"POST":   http.HandlerFunc(spah.createServerWithRandomPort),
		"DELETE": http.HandlerFunc(spah.removeServers),
	}).Name(routeServers)
	apiRouter.Handle("/servers/{port:[0-9]{1,5}}", handlers.MethodHandler{
		"GET":    http.HandlerFunc(spah.getServer),
		"PUT":    http.HandlerFunc(spah.createServer),
		"DELETE": http.HandlerFunc(spah.removeServer),
	}).Name(routeServersSelf)
	apiRouter.Handle("/servers/{port:[0-9]{1,5}}/aliases", handlers.MethodHandler{
		"GET":    http.HandlerFunc(spah.getServerAliases),
		"POST":   http.HandlerFunc(spah.createServerAlias),
		"DELETE": http.HandlerFunc(spah.removeServerAliases),
	}).Name(routeServersSelfAliases)
	apiRouter.Handle("/servers/{port:[0-9]{1,5}}/aliases/{alias}", handlers.MethodHandler{
		"GET":    http.HandlerFunc(spah.getServerAlias),
		"DELETE": http.HandlerFunc(spah.removeServerAlias),
	}).Name(routeServersSelfAliasesSelf)
	apiRouter.Handle("/fileservers", handlers.MethodHandler{
		"GET": http.HandlerFunc(spah.getFileServers),
	}).Name(routeFileservers)
	spah.h = r
	spah.router = r
	return &spah
}

func (spah *serverPoolAdminHandler) writeLocation(w http.ResponseWriter, name string, params ...string) {
	var urlStr string
	if u, err := spah.router.GetRoute(name).URL(params...); err != nil {
		log.Print(err)
		urlStr = ""
	} else {
		urlStr = u.String()
	}
	w.Header().Set("Location", urlStr)
}

func (spah *serverPoolAdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	spah.h.ServeHTTP(w, r)
}

type ServerData struct {
	BindAddress string      `json:"bind_address"`
	Port        uint16      `json:"port"`
	Aliases     []AliasData `json:"aliases"`
}

func newServerDataFromServer(srv *Server) *ServerData {
	aliasNames := srv.DirAliases.List()
	aliases := make([]AliasData, 0, len(aliasNames))
	for _, alias := range aliasNames {
		aliases = append(aliases, AliasData{
			ID:   aliasID(alias),
			Name: alias,
			Path: srv.DirAliases.Get(alias),
		})
	}
	host, _, _ := net.SplitHostPort(srv.Addr)
	return &ServerData{
		BindAddress: host,
		Port:        srv.Port,
		Aliases:     aliases,
	}
}

type AliasData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func aliasID(alias string) string {
	h := sha1.New()
	h.Write([]byte(alias))
	b := h.Sum(nil)
	return fmt.Sprintf("%x", b)[0:7]
}

func (spah *serverPoolAdminHandler) serverOr404(w http.ResponseWriter, r *http.Request) *Server {
	sport := mux.Vars(r)["port"]
	port, err := strconv.Atoi(sport)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}
	srv := spah.ServerPool.Get(uint16(port))
	if srv == nil {
		http.Error(w, fmt.Sprintf("server %d not found", port), http.StatusNotFound)
		return nil
	}
	return srv
}

func (spah *serverPoolAdminHandler) getServers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	srvsData := make([]ServerData, 0, len(spah.ServerPool.Srvs))
	for _, srv := range spah.ServerPool.Srvs {
		srvsData = append(srvsData, *newServerDataFromServer(srv))
	}
	if err := json.NewEncoder(w).Encode(srvsData); err != nil {
		log.Print(err)
	}
}

func (spah *serverPoolAdminHandler) getServer(w http.ResponseWriter, r *http.Request) {
	srv := spah.serverOr404(w, r)
	if srv == nil {
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(newServerDataFromServer(srv)); err != nil {
		log.Print(err)
	}
}

type CreateServerRequest struct {
	BindAddress string `json:"bind_address"`
}

func (spah *serverPoolAdminHandler) createServerWithRandomPort(w http.ResponseWriter, r *http.Request) {
	var req CreateServerRequest
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	addr := net.JoinHostPort(req.BindAddress, "0")
	srv, err := spah.ServerPool.Add(addr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Wait for the server to be started
	<-srv.started
	spah.writeLocation(w, routeServersSelf, "port", strconv.Itoa(int(srv.Port)))
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newServerDataFromServer(srv)); err != nil {
		log.Print(err)
	}
}

func (spah *serverPoolAdminHandler) createServer(w http.ResponseWriter, r *http.Request) {
	var req CreateServerRequest
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sport := mux.Vars(r)["port"]
	port, err := strconv.Atoi(sport)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	addr := net.JoinHostPort(req.BindAddress, strconv.Itoa(port))
	srv, err := spah.ServerPool.Add(addr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Wait for the server to be started
	<-srv.started
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(newServerDataFromServer(srv)); err != nil {
		log.Print(err)
	}
}

func (spah *serverPoolAdminHandler) removeServers(w http.ResponseWriter, r *http.Request) {
	var errs []error
	for _, srv := range spah.ServerPool.Srvs {
		if ok, err := spah.ServerPool.Remove(srv.Port); err != nil {
			errs = append(errs, err)
		} else if !ok {
			err := fmt.Errorf("error shutting down server on port %d", srv.Port)
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		var bufMsg bytes.Buffer
		for _, err := range errs {
			fmt.Fprintln(&bufMsg, err.Error())
		}
		apiErr := &AdminAPIError{apiErrTypeAPIInternal, bufMsg.String()}
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(apiErr); err != nil {
			log.Print(err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (spah *serverPoolAdminHandler) removeServer(w http.ResponseWriter, r *http.Request) {
	srv := spah.serverOr404(w, r)
	if srv == nil {
		return
	}
	if ok, err := spah.ServerPool.Remove(srv.Port); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		err := fmt.Errorf("error shutting down server on port %d", srv.Port)
		log.Print(err)
	}
	w.WriteHeader(http.StatusOK)
}

func (spah *serverPoolAdminHandler) getServerAliases(w http.ResponseWriter, r *http.Request) {
	srv := spah.serverOr404(w, r)
	if srv == nil {
		return
	}
	aliasNames := srv.DirAliases.List()
	aliases := make([]AliasData, 0, len(aliasNames))
	for _, alias := range aliasNames {
		aliases = append(aliases, AliasData{
			ID:   aliasID(alias),
			Name: alias,
			Path: srv.DirAliases.Get(alias),
		})
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(aliases); err != nil {
		log.Print(err)
	}
}

func (spah *serverPoolAdminHandler) removeServerAliases(w http.ResponseWriter, r *http.Request) {
	srv := spah.serverOr404(w, r)
	if srv == nil {
		return
	}
	aliases := srv.DirAliases.List()
	for _, alias := range aliases {
		srv.DirAliases.Delete(alias)
	}
	w.WriteHeader(http.StatusOK)
}

func (spah *serverPoolAdminHandler) getServerAlias(w http.ResponseWriter, r *http.Request) {
	srv := spah.serverOr404(w, r)
	if srv == nil {
		return
	}
	reqAliasID := mux.Vars(r)["alias"]
	var alias string
	for _, a := range srv.DirAliases.List() {
		if aliasID(a) == reqAliasID {
			alias = a
			break
		}
	}
	aliasPath := srv.DirAliases.Get(alias)
	if aliasPath == "" {
		http.Error(w, fmt.Sprintf("server %d has no alias %q", srv.Port, alias), http.StatusNotFound)
		return
	}

	aliasData := AliasData{
		ID:   aliasID(alias),
		Name: alias,
		Path: srv.DirAliases.Get(alias),
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(aliasData); err != nil {
		log.Print(err)
	}
}

type CreateServerAliasRequest struct {
	Alias    string            `json:"alias"`
	Path     string            `json:"path"`
	FsType   string            `json:"fs_type"`
	FsParams fileserver.Params `json:"fs_params"`
}

func (spah *serverPoolAdminHandler) createServerAlias(w http.ResponseWriter, r *http.Request) {
	srv := spah.serverOr404(w, r)
	if srv == nil {
		return
	}
	defer r.Body.Close()
	var req CreateServerAliasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := srv.DirAliases.Put(req.Alias, req.Path, req.FsType, req.FsParams); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	aliasData := AliasData{
		ID:   aliasID(req.Alias),
		Name: req.Alias,
		Path: srv.DirAliases.Get(req.Alias),
	}
	spah.writeLocation(w, routeServersSelfAliasesSelf, "alias", aliasData.ID)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(aliasData); err != nil {
		log.Print(err)
	}
}

func (spah *serverPoolAdminHandler) removeServerAlias(w http.ResponseWriter, r *http.Request) {
	srv := spah.serverOr404(w, r)
	if srv == nil {
		return
	}
	reqAliasID := mux.Vars(r)["alias"]
	var alias string
	for _, a := range srv.DirAliases.List() {
		if aliasID(a) == reqAliasID {
			alias = a
			break
		}
	}
	ok := srv.DirAliases.Delete(alias)
	if !ok {
		http.Error(w, fmt.Sprintf("server %d has no alias %q", srv.Port, alias), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (spah *serverPoolAdminHandler) getFileServers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(spah.ServerPool.fsf.Types()); err != nil {
		log.Print(err)
	}
}