package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const DefaultBaseURL = "http://localhost:11434"

const (
	probeTimeout = 2 * time.Second
	pullTimeout  = 30 * time.Minute
)

type APIStatus struct {
	BaseURL      string
	APIAvailable bool
	Error        string
	Models       []Model
}

type Model struct {
	Name    string       `json:"name"`
	Model   string       `json:"model"`
	Size    int64        `json:"size"`
	Details ModelDetails `json:"details"`
}

type ModelDetails struct {
	ParameterSize     string `json:"parameter_size"`
	QuantizationLevel string `json:"quantization_level"`
}

type PullRequest struct {
	Model    string `json:"model"`
	Insecure bool   `json:"insecure,omitempty"`
	Stream   bool   `json:"stream"`
}

type PullResponse struct {
	Status string `json:"status"`
}

type listModelsResponse struct {
	Models []Model `json:"models"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func Probe(ctx context.Context, baseURL string) APIStatus {
	models, err := ListModels(ctx, baseURL)
	if err != nil {
		return APIStatus{
			BaseURL:      normalizeBaseURL(baseURL),
			APIAvailable: false,
			Error:        err.Error(),
		}
	}

	return APIStatus{
		BaseURL:      normalizeBaseURL(baseURL),
		APIAvailable: true,
		Models:       models,
	}
}

func ListModels(ctx context.Context, baseURL string) ([]Model, error) {
	requestCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, apiURL(baseURL, "/tags"), nil)
	if err != nil {
		return nil, err
	}

	response, err := (&http.Client{Timeout: probeTimeout}).Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, decodeAPIError(response)
	}

	var payload listModelsResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}

	sort.Slice(payload.Models, func(i, j int) bool {
		return payload.Models[i].Name < payload.Models[j].Name
	})

	return payload.Models, nil
}

func PullModel(ctx context.Context, baseURL string, request PullRequest) (PullResponse, error) {
	request.Stream = false

	requestCtx, cancel := context.WithTimeout(ctx, pullTimeout)
	defer cancel()

	body, err := json.Marshal(request)
	if err != nil {
		return PullResponse{}, err
	}

	httpRequest, err := http.NewRequestWithContext(requestCtx, http.MethodPost, apiURL(baseURL, "/pull"), bytes.NewReader(body))
	if err != nil {
		return PullResponse{}, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := (&http.Client{Timeout: pullTimeout}).Do(httpRequest)
	if err != nil {
		return PullResponse{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return PullResponse{}, decodeAPIError(response)
	}

	var payload PullResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return PullResponse{}, err
	}

	return payload, nil
}

// ContextModelName names the derived model that runs base with a fixed context, e.g.
// qwen3.5:9b-q4_K_M with 32768 -> qwen3.5:9b-q4_K_M-ctx32k.
func ContextModelName(base string, numCtx int) string {
	if !strings.Contains(base, ":") {
		base += ":latest"
	}
	return fmt.Sprintf("%s-ctx%dk", base, numCtx/1024)
}

// CreateContextModel creates or updates ContextModelName(base, numCtx) with num_ctx and any extra
// parameters baked in. Ollama's OpenAI-compatible endpoint, which OpenCode uses, cannot set num_ctx per
// request. Creating from an installed model copies no weights, so repeating it is cheap and idempotent.
func CreateContextModel(ctx context.Context, baseURL, base string, numCtx int, extra map[string]any) error {
	parameters := map[string]any{"num_ctx": numCtx}
	for key, value := range extra {
		parameters[key] = value
	}
	body, err := json.Marshal(map[string]any{
		"model":      ContextModelName(base, numCtx),
		"from":       base,
		"parameters": parameters,
		"stream":     false,
	})
	if err != nil {
		return err
	}

	// Creating from an installed model only writes a manifest, so a long wait means Ollama is stuck.
	requestCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	httpRequest, err := http.NewRequestWithContext(requestCtx, http.MethodPost, apiURL(baseURL, "/create"), bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return decodeAPIError(response)
	}
	return nil
}

func apiURL(baseURL, path string) string {
	trimmed := strings.TrimRight(normalizeBaseURL(baseURL), "/")
	if strings.HasSuffix(trimmed, "/api") {
		return trimmed + path
	}
	return trimmed + "/api" + path
}

func normalizeBaseURL(baseURL string) string {
	if strings.TrimSpace(baseURL) == "" {
		return DefaultBaseURL
	}
	return strings.TrimRight(baseURL, "/")
}

func decodeAPIError(response *http.Response) error {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("unexpected status from Ollama API: %s", response.Status)
	}

	var apiError errorResponse
	if err := json.Unmarshal(body, &apiError); err == nil && apiError.Error != "" {
		return fmt.Errorf("%s", apiError.Error)
	}

	message := strings.TrimSpace(string(body))
	if message == "" {
		return fmt.Errorf("unexpected status from Ollama API: %s", response.Status)
	}

	return fmt.Errorf("%s", message)
}
