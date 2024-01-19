package protocol

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type AuthHandler interface {
	Auth(Credential, *http.Request) (bool, error)
}

type MiddlewareFunc = func(http.Handler) http.Handler

type HTTPServer struct {
	router      *chi.Mux
	AuthHandler AuthHandler
	config      Config
}

const (
	infoRefsSuffix      string = "/info/refs"
	gitUploadPackSuffix string = "/git-upload-pack"
	gitRecvPackSuffix   string = "/git-receive-pack"
)

func NewHTTPServer(config Config) HTTPServer {
	server := HTTPServer{config: config, router: chi.NewRouter()}

	if server.config.GitPath == "" {
		server.config.GitPath = "git"
	}

	return server
}

func (s *HTTPServer) Middlewares(handler http.Handler) http.Handler {
	return middleware.Logger(
		s.GetNamespaceAndRepo(
			s.Auth(
				handler,
			),
		),
	)
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path

	switch {
	case strings.HasSuffix(urlPath, infoRefsSuffix):
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		s.Middlewares(s.InfoRefs()).ServeHTTP(w, r)

	case strings.HasSuffix(urlPath, gitUploadPackSuffix):
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		s.Middlewares(s.PostRPC("git-upload-pack")).ServeHTTP(w, r)

	case strings.HasSuffix(urlPath, gitRecvPackSuffix):
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		s.Middlewares(s.PostRPC("git-receive-pack")).ServeHTTP(w, r)

	default:
		slog.Error("Unsupported url path: " + urlPath)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
}

func (s *HTTPServer) Setup() error {
	return s.config.Setup()
}

func (s *HTTPServer) Handler() http.Handler {
	return s.router
}
