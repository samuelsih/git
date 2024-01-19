package protocol

import (
	"log/slog"
	"net/http"
	"strings"
)

type AuthHandler interface {
	Auth(Credential, *http.Request) (bool, error)
}

type MiddlewareFunc = func(http.Handler) http.Handler

type HTTPServer struct {
	router      *http.ServeMux
	AuthHandler AuthHandler
	config      Config
}

const (
	infoRefsSuffix      string = "/info/refs"
	gitUploadPackSuffix string = "/git-upload-pack"
	gitRecvPackSuffix   string = "/git-receive-pack"
)

func NewHTTPServer(config Config) HTTPServer {
	server := HTTPServer{config: config, router: http.NewServeMux()}

	if server.config.GitPath == "" {
		server.config.GitPath = "git"
	}

	return server
}

func (s *HTTPServer) Middlewares(handler http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	if len(middlewares) < 1 {
		return handler
	}

	wrapped := handler

	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}

	return wrapped
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path

	middlewares := []MiddlewareFunc{
		s.Logger,
		s.GetNamespaceAndRepo,
		s.Auth,
	}

	switch {
	case strings.HasSuffix(urlPath, infoRefsSuffix):
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		s.Middlewares(s.InfoRefs(), middlewares...).ServeHTTP(w, r)

	case strings.HasSuffix(urlPath, gitUploadPackSuffix):
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		s.Middlewares(s.PostRPC("git-upload-pack"), middlewares...).ServeHTTP(w, r)

	case strings.HasSuffix(urlPath, gitRecvPackSuffix):
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		s.Middlewares(s.PostRPC("git-receive-pack"), middlewares...).ServeHTTP(w, r)

	default:
		slog.Error("Unsupported url path: " + urlPath)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
}

func (s *HTTPServer) Setup() error {
	return s.config.Setup()
}
