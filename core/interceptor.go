package core

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"resilienceshield/logs"
	"resilienceshield/storage"
)

type Interceptor struct {
	checker   *ConnectivityChecker
	remoteURL string
}

func NewInterceptor(checker *ConnectivityChecker, remoteURL string) *Interceptor {
	return &Interceptor{
		checker:   checker,
		remoteURL: remoteURL,
	}
}

// Handle traite chaque requête entrante
func (i *Interceptor) Handle(w http.ResponseWriter, r *http.Request) {
	// Lire le body de la requête
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Erreur lecture body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	endpoint := r.URL.Path
	method := r.Method

	if i.checker.IsOnline() {
		// ── MODE ONLINE : transférer vers le serveur distant ──
		logs.Info(fmt.Sprintf("ONLINE → %s %s", method, endpoint))
		i.forwardToRemote(w, r, method, endpoint, body)
	} else {
		// ── MODE OFFLINE : stocker dans SQLite ──
		logs.Warning(fmt.Sprintf("OFFLINE → %s %s stocké localement", method, endpoint))
		err := storage.Enqueue(method, endpoint, string(body))
		if err != nil {
			logs.Error("Erreur stockage: " + err.Error())
			http.Error(w, "Erreur stockage local", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, `{
			"status": "queued",
			"message": "Transaction stockée localement, sera synchronisée au retour du réseau",
			"pending": %d
		}`, storage.CountPending())
	}
}

// forwardToRemote transfère la requête vers le serveur distant
func (i *Interceptor) forwardToRemote(w http.ResponseWriter, r *http.Request,
	method, endpoint string, body []byte) {

	targetURL := i.remoteURL + endpoint

	req, err := http.NewRequest(method, targetURL, bytes.NewBuffer(body))
	if err != nil {
		logs.Error("Erreur création requête: " + err.Error())
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}

	// Copier les headers
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Échec transfert → stocker localement
		logs.Warning("Transfert échoué → stockage local")
		storage.Enqueue(method, endpoint, string(body))
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, `{"status": "queued", "message": "Réseau instable, transaction mise en attente"}`)
		return
	}
	defer resp.Body.Close()

	// Retransmettre la réponse au Mini ERP
	respBody, _ := io.ReadAll(resp.Body)
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)

	logs.Success(fmt.Sprintf("Réponse %d reçue de %s", resp.StatusCode, targetURL))
}
