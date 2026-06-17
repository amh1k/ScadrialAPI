package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"scadrialapi.abdulmoiz.net/internal/data/mocks"
)

func TestHealthCheck(t *testing.T) {
	userMock := mocks.UserModel{}
	permissionMock := mocks.PermissionModel{}
	app := NewTestApplication(t, &userMock, &permissionMock)
	app.config.env = "testing"
	ts := newTestServer(t, app.routes())
	defer ts.Close()
	client := ts.Client()
	resp, err := client.Get(ts.URL + "/v1/healthcheck")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("got status %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("got content-type %q, want %q", ct, "application/json")
	}
	var body struct {
		Status     string `json:"status"`
		SystemInfo struct {
			Environment string `json:"environment"`
			Version     string `json:"version"`
		} `json:"system_info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Status != "available" {
		t.Errorf("got status %q, want %q", body.Status, "available")
	}
	if body.SystemInfo.Environment != "testing" {
		t.Errorf("got environment %q, want %q", body.SystemInfo.Environment, "testing")
	}

}
