package main

import (
	"context"
	"encoding/json"
	"errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var kubernetesClient *kubernetes.Clientset

func main() {
	configFlags := genericclioptions.NewConfigFlags(true)
	restConfig, err := configFlags.ToRESTConfig()
	if err != nil {
		slog.Error("error creating REST config", "error", err)
		os.Exit(1)
	}
	kubernetesClient, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		slog.Error("error creating Kubernetes client", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr: ":80",
	}

	http.HandleFunc("/api/v1/getparams.execute", getParamsHandler)

	go func() {
		slog.Info("starting server", "port", 80)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", "error", err)
		}
		slog.Info("stopped serving new connections")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP shutdown error", "error", err)
	}
	slog.Info("graceful shutdown complete")
}

func getParamsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slog.Info("getting cluster config")
	clusterConfig, err := kubernetesClient.CoreV1().ConfigMaps("kube-system").Get(context.Background(), "cluster-config", v1.GetOptions{})
	if err != nil {
		slog.Error("error getting configmap", "error", err, "configmap", "kube-system/cluster-config")
		http.Error(w, "not able to get configmap", http.StatusInternalServerError)
		return
	}
	slog.Info("cluster config retrieved", "clusterConfig", clusterConfig)

	response := map[string]any{
		"output": map[string]any{
			"parameters": []any{
				clusterConfig.Data,
			},
		},
	}
	slog.Info("sending response", "response", response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error("error writing response", "error", err)
	}
}
