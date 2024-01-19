package protocol

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"
)

func (s *HTTPServer) GetNamespaceAndRepo(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path
		repoURLPath := ""

		slog.Info("URL Path: " + urlPath)

		switch {
		case strings.HasSuffix(urlPath, infoRefsSuffix):
			repoURLPath = strings.Replace(urlPath, infoRefsSuffix, "", 1)
			slog.Info("suffix: " + infoRefsSuffix)

		case strings.HasSuffix(urlPath, gitUploadPackSuffix):
			repoURLPath = strings.Replace(urlPath, gitUploadPackSuffix, "", 1)
			slog.Info("suffix: " + infoRefsSuffix)

		case strings.HasSuffix(urlPath, gitRecvPackSuffix):
			repoURLPath = strings.Replace(urlPath, gitRecvPackSuffix, "", 1)
			slog.Info("suffix: " + infoRefsSuffix)

		default:
			slog.Error("Unsupported url path: " + urlPath)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		slog.Info("GetNamespaceAndrepo | repoURLPath: " + repoURLPath)

		repoNamespace, repoName := getNamespaceAndRepo(repoURLPath)
		if repoName == "" {
			slog.Error("auth: no repo name provided")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		repoRequest := RepoContext{
			RepoName: path.Join(repoNamespace, repoName),
			RepoPath: path.Join(s.config.Dir, repoNamespace, repoName),
		}

		slog.Info(fmt.Sprintf("Repo request: %s - %s", repoRequest.RepoName, repoRequest.RepoPath))

		if !repoExists(repoRequest.RepoPath) && s.config.AutoCreate {
			err := initRepo(repoRequest.RepoName, &s.config)
			if err != nil {
				slog.Error("repo-init: " + err.Error())
			}
		}

		if !repoExists(repoRequest.RepoPath) {
			slog.Error("repo-init: " + fmt.Errorf("%s does not exist", repoRequest.RepoPath).Error())
			http.NotFound(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), repoContextKey, repoRequest)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(f)
}

func (s *HTTPServer) Auth(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Auth middleware started")

		if !s.config.Auth {
			slog.Warn("No auth provided, next...")
			next.ServeHTTP(w, r)
			return
		}

		slog.Info("Auth available, init authenticating")

		if s.AuthHandler.Auth == nil {
			slog.Error("auth: no auth backend provided")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header()["WWW-Authenticate"] = []string{`Basic realm=""`}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		cred, err := getCredential(r)
		if err != nil {
			slog.Error("auth: " + err.Error())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		allow, err := s.AuthHandler.Auth(cred, r)
		if !allow || err != nil {
			if err != nil {
				slog.Error("auth: " + err.Error())
			}

			slog.Error(fmt.Sprintf("rejected user %s", cred.Username))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	return http.HandlerFunc(f)
}

func (s *HTTPServer) Logger(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		defer func(start time.Time) {
			slog.Info(fmt.Sprintf("%s %s from %s - [%s] in %+v", r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent(), time.Since(start)))
		}(time.Now())

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}
