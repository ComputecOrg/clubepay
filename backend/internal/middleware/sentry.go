package middleware

import (
	"log/slog"
	"net/http"

	"github.com/getsentry/sentry-go"
)

// SentryRecovery retorna um middleware que captura panics e os reporta ao Sentry (se dsn != "").
// Quando dsn está vazio, apenas loga e responde 500 sem enviar ao Sentry.
func SentryRecovery(dsn string) func(http.Handler) http.Handler {
	if dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			TracesSampleRate: 0.1,
		}); err != nil {
			slog.Warn("falha ao inicializar Sentry", "error", err)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					slog.Error("panic recuperado", "error", rec, "path", r.URL.Path)

					if dsn != "" {
						hub := sentry.CurrentHub().Clone()
						hub.Recover(rec)
						hub.Flush(0)
					}

					http.Error(w, `{"error":"erro interno"}`, http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
