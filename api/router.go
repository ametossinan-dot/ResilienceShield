package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"resilienceshield/core"
	"resilienceshield/logs"
	"resilienceshield/storage"

	"github.com/gorilla/mux"
)

type Router struct {
	checker     *core.ConnectivityChecker
	interceptor *core.Interceptor
	syncEngine  *core.SyncEngine
}

func NewRouter(
	checker *core.ConnectivityChecker,
	interceptor *core.Interceptor,
	syncEngine *core.SyncEngine,
) *Router {
	return &Router{checker, interceptor, syncEngine}
}

func (ro *Router) Setup() *mux.Router {
	r := mux.NewRouter()

	// ── Endpoints du middleware ──────────────────────────────
	r.HandleFunc("/status", ro.getStatus).Methods("GET")
	r.HandleFunc("/sync", ro.triggerSync).Methods("POST")
	r.HandleFunc("/queue", ro.getQueue).Methods("GET")
	r.HandleFunc("/logs", ro.getLogs).Methods("GET")

	// ── Interface admin (fichiers statiques) ─────────────────
	r.PathPrefix("/admin/").Handler(
		http.StripPrefix("/admin/",
			http.FileServer(http.Dir("./admin/"))))

	// ── Intercepteur : toutes les autres requêtes ────────────
	r.PathPrefix("/").HandlerFunc(ro.interceptor.Handle)

	return r
}

// GET /status
func (ro *Router) getStatus(w http.ResponseWriter, r *http.Request) {
	status := "OFFLINE"
	if ro.checker.IsOnline() {
		status = "ONLINE"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  status,
		"pending": storage.CountPending(),
	})
}

// POST /sync
func (ro *Router) triggerSync(w http.ResponseWriter, r *http.Request) {
	if !ro.checker.IsOnline() {
		http.Error(w, `{"error":"Pas de connexion"}`, http.StatusServiceUnavailable)
		return
	}
	go ro.syncEngine.Sync()
	logs.Info("Synchronisation manuelle déclenchée.")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message":"Synchronisation lancée"}`)
}

// GET /queue
func (ro *Router) getQueue(w http.ResponseWriter, r *http.Request) {
	pending, err := storage.GetPending()
	if err != nil {
		http.Error(w, "Erreur", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pending)
}

// GET /logs
func (ro *Router) getLogs(w http.ResponseWriter, r *http.Request) {
	rows, err := storage.DB.Query(
		`SELECT message, created_at FROM sync_logs ORDER BY created_at DESC LIMIT 50`,
	)
	if err != nil {
		http.Error(w, "Erreur", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []map[string]string
	for rows.Next() {
		var msg, date string
		rows.Scan(&msg, &date)
		result = append(result, map[string]string{
			"message": msg, "date": date,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
