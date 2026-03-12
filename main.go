package main

import (
	"fmt"
	"net/http"
	"resilienceshield/api"
	"resilienceshield/config"
	"resilienceshield/core"
	"resilienceshield/logs"
	"resilienceshield/storage"
)

func main() {
	// Charger la configuration
	cfg := config.Load()

	// Initialiser les logs
	logs.Init()
	logs.Info("Démarrage de ResilienceShield Middleware…")

	// Initialiser la base SQLite
	storage.InitDB(cfg.LocalDBPath)

	// Démarrer le ConnectivityChecker
	checker := core.NewConnectivityChecker(cfg.RemoteHost, cfg.CheckInterval)

	// Démarrer le SyncEngine
	syncEngine := core.NewSyncEngine(cfg.RemoteDBUrl)

	// Quand le réseau revient → sync automatique
	checker.OnStatusChange = func(online bool) {
		if online {
			logs.Info("Déclenchement sync automatique…")
			go syncEngine.Sync()
		}
	}

	checker.Start()
	syncEngine.Start(checker)

	// Créer l'intercepteur
	interceptor := core.NewInterceptor(checker, cfg.RemoteDBUrl)

	// Configurer le routeur
	router := api.NewRouter(checker, interceptor, syncEngine)
	r := router.Setup()

	// Démarrer le serveur
	addr := fmt.Sprintf(":%s", cfg.Port)
	logs.Info(fmt.Sprintf("Middleware démarré sur http://localhost%s", addr))
	logs.Info("Dashboard admin : http://localhost:8080/admin/")

	if err := http.ListenAndServe(addr, r); err != nil {
		logs.Error("Erreur serveur : " + err.Error())
	}
}
