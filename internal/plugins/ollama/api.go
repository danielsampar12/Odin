package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

const DefaultBaseURL = "http://localhost:11434"

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

type listModelsResponse struct {
	Models []Model `json:"models"`
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
	requestCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, apiURL(baseURL, "/tags"), nil)
	if err != nil {
		return nil, err
	}

	response, err := (&http.Client{Timeout: 2 * time.Second}).Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status from Ollama API: %s", response.Status)
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
