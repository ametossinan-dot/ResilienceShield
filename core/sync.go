package core

import (
	"bytes"
	"fmt"
	"net/http"
	"resilienceshield/logs"
	"resilienceshield/storage"
	"time"
)

type SyncEngine struct {
	remoteURL string
	running   bool
}

func NewSyncEngine(remoteURL string) *SyncEngine {
	return &SyncEngine{remoteURL: remoteURL}
}

// Start lance la synchronisation automatique
func (s *SyncEngine) Start(checker *ConnectivityChecker) {
	go func() {
		for {
			time.Sleep(10 * time.Second)
			if checker.IsOnline() && storage.CountPending() > 0 {
				s.Sync()
			}
		}
	}()
}

// Sync synchronise toutes les transactions en attente
func (s *SyncEngine) Sync() {
	pending, err := storage.GetPending()
	if err != nil {
		logs.Error("Erreur lecture file FIFO: " + err.Error())
		return
	}

	if len(pending) == 0 {
		logs.Info("Aucune transaction en attente.")
		return
	}

	logs.Info(fmt.Sprintf("Synchronisation de %d transaction(s)…", len(pending)))

	success := 0
	failed := 0

	for _, t := range pending {
		err := s.sendTransaction(t)
		if err != nil {
			logs.Error(fmt.Sprintf("Échec transaction #%d : %s", t.ID, err.Error()))
			failed++
			continue
		}
		storage.MarkSynced(t.ID)
		logs.Success(fmt.Sprintf("Transaction #%d synchronisée.", t.ID))
		success++
	}

	logs.Info(fmt.Sprintf("Sync terminée — %d succès / %d échecs", success, failed))

	// Enregistrer dans sync_logs
	storage.DB.Exec(
		`INSERT INTO sync_logs (message) VALUES (?)`,
		fmt.Sprintf("Sync: %d succès, %d échecs", success, failed),
	)
}

func (s *SyncEngine) sendTransaction(t storage.Transaction) error {
	url := s.remoteURL + t.Endpoint
	req, err := http.NewRequest(t.Method, url, bytes.NewBufferString(t.Body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sync-Source", "ResilienceShield")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("serveur a retourné %d", resp.StatusCode)
	}
	return nil
}
