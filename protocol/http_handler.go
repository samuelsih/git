package protocol

import (
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type RepoContext struct {
	RepoName string
	RepoPath string
}

const repoContextKey = "repo-context"

func (s *HTTPServer) InfoRefs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Starting info refs")

		const ctx = "get-info-refs"
		repoCtx := r.Context().Value(repoContextKey).(RepoContext)
		if repoCtx == (RepoContext{}) {
			http.Error(w, "Unknown repo name and repo path", http.StatusBadRequest)
			return
		}

		rpc := r.URL.Query().Get("service")

		if !(rpc == "git-upload-pack" || rpc == "git-receive-pack") {
			http.NotFound(w, r)
			return
		}

		cmd, pipe := gitCommand(
			s.config.GitPath,
			subCommand(rpc),
			"--stateless-rpc", "--advertise-refs",
			repoCtx.RepoPath,
		)

		if err := cmd.Start(); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("InfoRefs | cmd.Start | %s: %v\n", ctx, err))
			return
		}

		defer cleanUpProcessGroup(cmd)

		w.Header().Add("Content-Type", fmt.Sprintf("application/x-%s-advertisement", rpc))
		w.Header().Add("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)

		if err := packLine(w, fmt.Sprintf("# service=%s\n", rpc)); err != nil {
			slog.Error(fmt.Sprintf("InfoRefs | packLine | %s: %v\n", ctx, err))
			return
		}

		if err := packFlush(w); err != nil {
			slog.Error(fmt.Sprintf("InfoRefs | packFlush | %s: %v\n", ctx, err))
			return
		}

		rc := http.NewResponseController(w)
		defer rc.Flush()

		if _, err := io.Copy(w, pipe); err != nil {
			slog.Error(fmt.Sprintf("InfoRefs | io.Copy | %s: %v\n", ctx, err))
			return
		}

		if err := cmd.Wait(); err != nil {
			slog.Error(fmt.Sprintf("InfoRefs | cmd.Wait | %s: %v\n", ctx, err))
			return
		}
	}
}

func (s *HTTPServer) PostRPC(rpc string) http.HandlerFunc {
	const ctx = "post-rpc"

	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Starting post rpc")

		body := r.Body
		repoCtx := r.Context().Value(repoContextKey).(RepoContext)
		if repoCtx == (RepoContext{}) {
			http.Error(w, "Unknown repo name and repo path", http.StatusBadRequest)
			return
		}

		if r.Header.Get("Content-Encoding") == "gzip" {
			var err error
			body, err = gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				slog.Error(fmt.Sprintf("Gzip error | %s: %v\n", ctx, err))
				return
			}
		}

		cmd, pipe := gitCommand(
			s.config.GitPath,
			subCommand(rpc),
			"--stateless-rpc",
			repoCtx.RepoPath,
		)

		defer pipe.Close()

		stdin, err := cmd.StdinPipe()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("StdinPipe | %s: %v\n", ctx, err))
			return
		}

		defer stdin.Close()

		if err := cmd.Start(); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("Cmd | %s: %v\n", ctx, err))
			return
		}

		defer cleanUpProcessGroup(cmd)

		if _, err := io.Copy(stdin, body); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			slog.Error(fmt.Sprintf("io.Copy | %s: %v\n", ctx, err))
			return
		}

		stdin.Close()

		w.Header().Add("Content-Type", fmt.Sprintf("application/x-%s-result", rpc))
		w.Header().Add("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)

		if _, err := io.Copy(w, pipe); err != nil {
			slog.Error(fmt.Sprintf("io.Copy Write Flusher | %s: %s", ctx, err.Error()))
			return
		}

		if err := cmd.Wait(); err != nil {
			slog.Error(fmt.Sprintf("cmd.Wait | %s: %s", ctx, err.Error()))
			return
		}
	}
}
