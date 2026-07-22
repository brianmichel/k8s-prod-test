package main

import (
	"crypto/tls"
	"crypto/x509"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	iofs "io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

//go:embed web/*
var web embed.FS

type kubeClient struct {
	baseURL string
	token   string
	client  *http.Client
}

type resourceList struct {
	Items []map[string]any `json:"items"`
}

type overview struct {
	GeneratedAt  time.Time        `json:"generatedAt"`
	Applications []map[string]any `json:"applications"`
	Rollouts     []map[string]any `json:"rollouts"`
	Deployments  []map[string]any `json:"deployments"`
	ReplicaSets  []map[string]any `json:"replicaSets"`
	Services     []map[string]any `json:"services"`
	AnalysisRuns []map[string]any `json:"analysisRuns"`
	Pods         []map[string]any `json:"pods"`
	Warnings     []string         `json:"warnings,omitempty"`
}

func newKubeClient() (*kubeClient, error) {
	host := env("KUBERNETES_SERVICE_HOST", "kubernetes.default.svc")
	port := env("KUBERNETES_SERVICE_PORT_HTTPS", "443")
	token, err := os.ReadFile(env("SERVICE_ACCOUNT_TOKEN_PATH", "/var/run/secrets/kubernetes.io/serviceaccount/token"))
	if err != nil {
		return nil, fmt.Errorf("read service account token: %w", err)
	}
	ca, err := os.ReadFile(env("SERVICE_ACCOUNT_CA_PATH", "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"))
	if err != nil {
		return nil, fmt.Errorf("read cluster CA: %w", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(ca) {
		return nil, errors.New("cluster CA contains no certificates")
	}
	return &kubeClient{
		baseURL: "https://" + host + ":" + port,
		token:   strings.TrimSpace(string(token)),
		client: &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			RootCAs:    roots,
		}}},
	}, nil
}

func (k *kubeClient) request(method, resourcePath string, body io.Reader, contentType string) ([]byte, error) {
	req, err := http.NewRequest(method, k.baseURL+resourcePath, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+k.token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := k.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("kubernetes API returned %s: %s", resp.Status, strings.TrimSpace(string(payload)))
	}
	return payload, nil
}

func (k *kubeClient) list(resourcePath string) ([]map[string]any, error) {
	payload, err := k.request(http.MethodGet, resourcePath, nil, "")
	if err != nil {
		return nil, err
	}
	var list resourceList
	if err := json.Unmarshal(payload, &list); err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (k *kubeClient) patch(resourcePath, patch string) error {
	_, err := k.request(http.MethodPatch, resourcePath, strings.NewReader(patch), "application/merge-patch+json")
	return err
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	kube, err := newKubeClient()
	if err != nil {
		logger.Error("cannot initialize Kubernetes client", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	mux.HandleFunc("GET /api/overview", overviewHandler(kube, logger))
	mux.HandleFunc("POST /api/applications/{namespace}/{name}/sync", actionHandler(kube, logger, "sync"))
	mux.HandleFunc("POST /api/rollouts/{namespace}/{name}/promote", actionHandler(kube, logger, "promote"))
	mux.HandleFunc("POST /api/rollouts/{namespace}/{name}/abort", actionHandler(kube, logger, "abort"))
	mux.Handle("/", http.FileServer(http.FS(mustSub(web, "web"))))

	server := &http.Server{Addr: ":8080", Handler: securityHeaders(mux), ReadHeaderTimeout: 5 * time.Second}
	logger.Info("flightdeck ready", "address", server.Addr)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func overviewHandler(kube *kubeClient, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		result := overview{GeneratedAt: time.Now().UTC()}
		queries := []struct {
			label, path string
			target      *[]map[string]any
		}{
			{"applications", "/apis/argoproj.io/v1alpha1/applications", &result.Applications},
			{"rollouts", "/apis/argoproj.io/v1alpha1/rollouts", &result.Rollouts},
			{"deployments", "/apis/apps/v1/deployments", &result.Deployments},
			{"replica sets", "/apis/apps/v1/replicasets", &result.ReplicaSets},
			{"services", "/api/v1/services", &result.Services},
			{"analysis runs", "/apis/argoproj.io/v1alpha1/analysisruns", &result.AnalysisRuns},
			{"pods", "/api/v1/pods", &result.Pods},
		}
		for _, query := range queries {
			items, err := kube.list(query.path)
			if err != nil {
				logger.Warn("resource query failed", "resource", query.label, "error", err)
				result.Warnings = append(result.Warnings, query.label+": unavailable")
				continue
			}
			*query.target = items
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func actionHandler(kube *kubeClient, logger *slog.Logger, action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace, name := path.Clean(r.PathValue("namespace")), path.Clean(r.PathValue("name"))
		if namespace == "." || name == "." || strings.Contains(namespace, "/") || strings.Contains(name, "/") {
			writeError(w, http.StatusBadRequest, "invalid resource name")
			return
		}
		var resourcePath, patch string
		switch action {
		case "sync":
			resourcePath = fmt.Sprintf("/apis/argoproj.io/v1alpha1/namespaces/%s/applications/%s", namespace, name)
			patch = `{"operation":{"initiatedBy":{"username":"flightdeck"},"sync":{"prune":true}}}`
		case "promote":
			resourcePath = fmt.Sprintf("/apis/argoproj.io/v1alpha1/namespaces/%s/rollouts/%s/status", namespace, name)
			patch = `{"status":{"promoteFull":true}}`
		case "abort":
			resourcePath = fmt.Sprintf("/apis/argoproj.io/v1alpha1/namespaces/%s/rollouts/%s", namespace, name)
			patch = `{"spec":{"abort":true}}`
		}
		if err := kube.patch(resourcePath, patch); err != nil {
			logger.Error("action failed", "action", action, "resource", namespace+"/"+name, "error", err)
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		logger.Info("action requested", "action", action, "resource", namespace+"/"+name)
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted", "action": action})
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self'; font-src 'self'; connect-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func mustSub(files embed.FS, dir string) iofs.FS {
	sub, err := iofs.Sub(files, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
