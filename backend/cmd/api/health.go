package main

import (
	"context"
	"encoding/json"
	"net/http"
)

// DBPinger é uma interface para verificar conectividade com o banco de dados.
type DBPinger interface {
	Ping(ctx context.Context) error
}

// healthzHandler retorna um http.Handler que verifica o status real da aplicação.
// Responde 200 quando tudo está saudável, 503 quando o banco está inacessível.
func healthzHandler(db DBPinger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := db.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck
				"status": "degraded",
				"db":     "error",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck
			"status": "ok",
			"db":     "ok",
		})
	})
}
