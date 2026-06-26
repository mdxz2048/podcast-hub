package http

import (
	"context"
	stdhttp "net/http"
	"net/url"
	"strings"

	"github.com/mdxz2048/podcast-hub/internal/auth"
)

type contextKey string

const authUserContextKey contextKey = "auth_user"

func (s *Server) corsMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == s.cfg.FrontendOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		}
		if r.Method == stdhttp.MethodOptions {
			w.WriteHeader(stdhttp.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) csrfOriginMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if r.Method != stdhttp.MethodPost && r.Method != stdhttp.MethodPut && r.Method != stdhttp.MethodPatch && r.Method != stdhttp.MethodDelete {
			next.ServeHTTP(w, r)
			return
		}
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin == "" {
			referer := strings.TrimSpace(r.Header.Get("Referer"))
			if referer == "" {
				writeError(w, r, stdhttp.StatusForbidden, "csrf_invalid", "请求未通过来源校验。")
				return
			}
			parsed, err := url.Parse(referer)
			if err != nil || parsed.Scheme == "" || parsed.Host == "" {
				writeError(w, r, stdhttp.StatusForbidden, "csrf_invalid", "请求未通过来源校验。")
				return
			}
			origin = parsed.Scheme + "://" + parsed.Host
		}
		if origin != s.cfg.FrontendOrigin {
			writeError(w, r, stdhttp.StatusForbidden, "csrf_invalid", "请求未通过来源校验。")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) securityHeadersMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) validateCSRFFromCookie(r *stdhttp.Request) error {
	cookie, err := r.Cookie("podcast_hub_csrf")
	if err != nil {
		return auth.ErrForbiddenCSRF
	}
	header := strings.TrimSpace(r.Header.Get(s.cfg.CSRFHeaderName))
	if header == "" || cookie.Value == "" || header != cookie.Value {
		return auth.ErrForbiddenCSRF
	}
	return nil
}

func (s *Server) RequireAuth(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		sessionCookie, err := r.Cookie(s.cfg.SessionCookieName)
		if err != nil {
			s.writeAuthError(w, r, auth.ErrNotAuthenticated)
			return
		}
		_, user, err := s.resolveSession(r.Context(), sessionCookie.Value)
		if err != nil {
			s.writeAuthError(w, r, err)
			return
		}
		ctx := context.WithValue(r.Context(), authUserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) RequireAdmin(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		user, ok := authUserFromContext(r.Context())
		if !ok {
			s.writeAuthError(w, r, auth.ErrNotAuthenticated)
			return
		}
		if user.Role != auth.RoleAdmin {
			writeError(w, r, stdhttp.StatusForbidden, "forbidden", "当前账号无权限访问该资源。")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func authUserFromContext(ctx context.Context) (auth.User, bool) {
	user, ok := ctx.Value(authUserContextKey).(auth.User)
	if !ok {
		return auth.User{}, false
	}
	return user, true
}
