package http

import (
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type statusRecorder struct {
	stdhttp.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = stdhttp.StatusOK
	}
	return r.ResponseWriter.Write(body)
}

func (s *Server) requestLogMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		started := time.Now()
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		status := rec.status
		if status == 0 {
			status = stdhttp.StatusOK
		}
		slog.Info("http_request",
			"request_id", middleware.GetReqID(r.Context()),
			"method", r.Method,
			"path", redactPrivateRSSPath(r.URL.Path),
			"status", status,
			"duration_ms", time.Since(started).Milliseconds(),
		)
	})
}

func (s *Server) handleMetrics(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	ctx := r.Context()
	dbReady := metricReadyValue(s.pingDatabase(ctx))
	redisStatus := s.pingRedis(ctx)
	redisReady := metricReadyValue(redisStatus)
	runnerEnabled := 0
	if s.cfg.RunnerMode == "docker_trusted_admin" {
		runnerEnabled = 1
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(stdhttp.StatusOK)
	_, _ = fmt.Fprintf(w, "# HELP podcast_hub_api_up API process liveness.\n")
	_, _ = fmt.Fprintf(w, "# TYPE podcast_hub_api_up gauge\n")
	_, _ = fmt.Fprintf(w, "podcast_hub_api_up 1\n")
	_, _ = fmt.Fprintf(w, "# HELP podcast_hub_dependency_ready Redacted dependency readiness.\n")
	_, _ = fmt.Fprintf(w, "# TYPE podcast_hub_dependency_ready gauge\n")
	_, _ = fmt.Fprintf(w, "podcast_hub_dependency_ready{dependency=\"database\"} %d\n", dbReady)
	_, _ = fmt.Fprintf(w, "podcast_hub_dependency_ready{dependency=\"redis\"} %d\n", redisReady)
	_, _ = fmt.Fprintf(w, "# HELP podcast_hub_runner_enabled Whether API configuration allows the separate runner mode.\n")
	_, _ = fmt.Fprintf(w, "# TYPE podcast_hub_runner_enabled gauge\n")
	_, _ = fmt.Fprintf(w, "podcast_hub_runner_enabled %d\n", runnerEnabled)
}

func metricReadyValue(status string) int {
	if status == "ok" || status == "unknown" {
		return 1
	}
	return 0
}

func redactPrivateRSSPath(path string) string {
	if !strings.HasPrefix(path, "/rss/private/") {
		return path
	}
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		return "/rss/private/[redacted]"
	}
	if strings.HasSuffix(parts[3], ".xml") {
		parts[3] = "[redacted].xml"
	} else {
		parts[3] = "[redacted]"
	}
	return strings.Join(parts, "/")
}
