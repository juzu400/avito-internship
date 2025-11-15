//go:build integration

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/juzu400/avito-internship/internal/logger"
	"github.com/juzu400/avito-internship/internal/migrations"
	"github.com/juzu400/avito-internship/internal/repository"
	"github.com/juzu400/avito-internship/internal/service"
	httptransport "github.com/juzu400/avito-internship/internal/transport/http"
)

func newTestServer(t *testing.T) (*httptest.Server, *repository.DB) {
	t.Helper()

	log := logger.New(logger.Config{Level: "debug"})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	const testDSN = "postgres://avito:avito@localhost:5433/avito_test?sslmode=disable"

	db, err := repository.NewPostgresDB(ctx, testDSN)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	t.Cleanup(db.Close)

	migrationsDir := migrationsPath(t)
	if err := migrations.Apply(ctx, log, db.Pool, migrationsDir); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	if _, err := db.Pool.Exec(ctx, `
        TRUNCATE TABLE pull_request_reviewers, pull_requests, team_members, teams, users
        RESTART IDENTITY CASCADE;
    `); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	repos := repository.NewRepositories(db)
	services := service.NewServices(log, repos)
	router := httptransport.NewRouter(log, services)

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	return srv, db
}

func doJSON(t *testing.T, client *http.Client, method, url string, body any) *http.Response {
	t.Helper()

	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		rdr = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func TestE2E_CreatePRAndStats(t *testing.T) {
	srv, _ := newTestServer(t)

	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := srv.URL

	// 1. Создаём команду с двумя пользователями
	teamReq := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}

	resp := doJSON(t, client, http.MethodPost, baseURL+"/team/add", teamReq)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("add team: unexpected status %d, body: %s", resp.StatusCode, string(b))
	}

	// 2. Создаём PR для этой команды
	createPRReq := map[string]any{
		"pull_request_id":   "pr-e2e-1",
		"pull_request_name": "E2E PR",
		"author_id":         "u1",
		"team_name":         "backend",
	}

	resp = doJSON(t, client, http.MethodPost, baseURL+"/pullRequest/create", createPRReq)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("create PR: unexpected status %d, body: %s", resp.StatusCode, string(b))
	}

	// 3. Проверяем статистику по пользователям-ревьюерам
	resp = doJSON(t, client, http.MethodGet, baseURL+"/users/stats", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("users stats: unexpected status %d, body: %s", resp.StatusCode, string(b))
	}

	var userStats struct {
		Items []struct {
			ReviewerID  string `json:"reviewer_id"`
			Assignments int    `json:"assignments"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userStats); err != nil {
		t.Fatalf("decode users stats: %v", err)
	}

	var foundU2 bool
	for _, it := range userStats.Items {
		if it.ReviewerID == "u2" {
			foundU2 = true
			if it.Assignments != 1 {
				t.Fatalf("expected assignments=1 for u2, got %d", it.Assignments)
			}
		}
	}
	if !foundU2 {
		t.Fatalf("expected to find stats for user u2")
	}

	// 4. Проверяем статистику по PR
	resp = doJSON(t, client, http.MethodGet, baseURL+"/pullRequests/stats", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("pullRequests stats: unexpected status %d, body: %s", resp.StatusCode, string(b))
	}

	var prStats struct {
		Items []struct {
			PullRequestID string `json:"pull_request_id"`
			Reviewers     int    `json:"reviewers"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&prStats); err != nil {
		t.Fatalf("decode pr stats: %v", err)
	}

	var foundPR bool
	for _, it := range prStats.Items {
		if it.PullRequestID == "pr-e2e-1" {
			foundPR = true
			if it.Reviewers != 1 {
				t.Fatalf("expected reviewers=1 for pr-e2e-1, got %d", it.Reviewers)
			}
		}
	}
	if !foundPR {
		t.Fatalf("expected to find pr-e2e-1 in pullRequests stats")
	}
}

func migrationsPath(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}

	dir := filepath.Dir(filename)
	root := filepath.Join(dir, "..", "..")
	return filepath.Join(root, "migrations")
}
