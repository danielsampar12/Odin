package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", request.Method)
		}
		if request.URL.Path != "/api/tags" {
			t.Fatalf("path = %s, want /api/tags", request.URL.Path)
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"models": []map[string]any{
				{
					"name": "qwen2.5-coder:7b",
					"details": map[string]any{
						"parameter_size":     "7B",
						"quantization_level": "Q4_K_M",
					},
				},
				{
					"name": "qwen2.5-coder:14b",
					"details": map[string]any{
						"parameter_size":     "14B",
						"quantization_level": "Q5_K_M",
					},
				},
			},
		})
	}))
	defer server.Close()

	models, err := ListModels(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("ListModels error = %v", err)
	}

	if len(models) != 2 {
		t.Fatalf("len(models) = %d, want 2", len(models))
	}
	if models[0].Name != "qwen2.5-coder:14b" {
		t.Fatalf("models[0].Name = %q, want sorted models", models[0].Name)
	}
}

func TestPullModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", request.Method)
		}
		if request.URL.Path != "/api/pull" {
			t.Fatalf("path = %s, want /api/pull", request.URL.Path)
		}

		var pullRequest PullRequest
		if err := json.NewDecoder(request.Body).Decode(&pullRequest); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if pullRequest.Model != "qwen2.5-coder:7b" {
			t.Fatalf("model = %q, want qwen2.5-coder:7b", pullRequest.Model)
		}
		if pullRequest.Stream {
			t.Fatalf("stream = true, want false")
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(PullResponse{Status: "success"})
	}))
	defer server.Close()

	response, err := PullModel(context.Background(), server.URL, PullRequest{Model: "qwen2.5-coder:7b"})
	if err != nil {
		t.Fatalf("PullModel error = %v", err)
	}
	if response.Status != "success" {
		t.Fatalf("status = %q, want success", response.Status)
	}
}

func TestPullModelError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(writer).Encode(map[string]string{
			"error": "model not found",
		})
	}))
	defer server.Close()

	_, err := PullModel(context.Background(), server.URL, PullRequest{Model: "missing-model"})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "model not found" {
		t.Fatalf("error = %q, want %q", err.Error(), "model not found")
	}
}
