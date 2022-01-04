package main

import (
	"net/http"

	"github.com/dimiro1/health"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func serve() {
	// Variables
	address := context.Configuration.Address
	healthHandler := health.NewHandler()
	// Listen path
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/health", healthHandler)
	// Listen
	err := http.ListenAndServe(address, nil)
	if err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}

	logrus.Info("Server listening on address " + address)
}
