package obligacje_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const (
	imageName     = "obligacje-integration-test"
	containerName = "obligacje-integration-test"
	serverPort    = "18080"
	serverAddr    = "http://localhost:" + serverPort
)

type valuationResponse struct {
	Name        string  `json:"name"`
	ISIN        string  `json:"isin"`
	ValuatedAt  string  `json:"valuated_at"`
	PurchaseDay int     `json:"purchase_day"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
}

func TestDockerIntegration_BondValuation(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not found in PATH, skipping integration test")
	}

	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	t.Log("Building Docker image…")
	build := exec.Command("docker", "build", "-t", imageName, ".")
	build.Dir = repoRoot
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("docker build failed: %v", err)
	}

	exec.Command("docker", "rm", "-f", containerName).Run() //nolint:errcheck

	t.Log("Starting container…")
	run := exec.Command(
		"docker", "run",
		"--name", containerName,
		"-p", serverPort+":8080",
		"-d",
		imageName,
	)
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	if err := run.Run(); err != nil {
		t.Fatalf("docker run failed: %v", err)
	}

	t.Cleanup(func() {
		if t.Failed() {
			logs, _ := exec.Command("docker", "logs", containerName).CombinedOutput()
			t.Logf("container logs:\n%s", logs)
		}
		t.Log("Removing container…")
		exec.Command("docker", "rm", "-f", containerName).Run() //nolint:errcheck
	})

	waitTime := 3 * time.Minute
	t.Logf("Waiting for server to be ready (up to %s, server downloads bond data on startup)…", waitTime)
	if err := waitForServer(serverAddr+"/v1/bond/TOS112501/valuation?valuated_at=2023-03-26", waitTime); err != nil {
		logs, _ := exec.Command("docker", "logs", containerName).CombinedOutput()
		t.Logf("container logs:\n%s", logs)
		t.Fatalf("server did not become ready: %v", err)
	}

	t.Run("TOS bond plain text valuation", func(t *testing.T) {
		url := serverAddr + "/v1/bond/TOS112501/valuation?valuated_at=2023-03-26"
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Accept", "text/plain")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		got := strings.TrimSpace(string(body))
		want := "102.72"
		if got != want {
			t.Errorf("got price %q, want %q", got, want)
		}
	})

	t.Run("TOS bond JSON valuation", func(t *testing.T) {
		url := serverAddr + "/v1/bond/TOS112501/valuation?valuated_at=2023-03-26"
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status %d", resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected Content-Type %q", ct)
		}

		var body valuationResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		if body.Name != "TOS112501" {
			t.Errorf("got name %q, want TOS112501", body.Name)
		}
		if body.ValuatedAt != "2023-03-26" {
			t.Errorf("got valuated_at %q, want 2023-03-26", body.ValuatedAt)
		}
		if body.Price != 102.72 {
			t.Errorf("got price %.2f, want 102.72", body.Price)
		}
		if body.Currency != "PLN" {
			t.Errorf("got currency %q, want PLN", body.Currency)
		}
		if body.ISIN == "" {
			t.Error("expected non-empty ISIN")
		}
	})
}

// waitForServer polls the given URL until it returns HTTP 200 or the deadline
// is reached. It uses an exponential backoff starting at 2 s up to 10 s.
func waitForServer(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	delay := 2 * time.Second
	for time.Now().Before(deadline) {
		resp, err := http.Get(url) //nolint:noctx
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(delay)
		if delay < 10*time.Second {
			delay = min(delay*2, 10*time.Second)
		}
	}
	return fmt.Errorf("server not ready after %s", timeout)
}
