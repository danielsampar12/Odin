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
