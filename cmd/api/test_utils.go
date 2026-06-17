package main

import (
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"scadrialapi.abdulmoiz.net/internal/data"
	"scadrialapi.abdulmoiz.net/internal/mailer"
)
func NewTestApplication(t *testing.T, mockUsers data.UserModelInterface, mockPermissions data.PermissionModelInterface) *application {
	cfg := config{}
	cfg.limiter.enabled = false

	return &application{
		config: cfg,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		models: data.Models{
			Users: mockUsers,
			Permissions: mockPermissions,
		},
		mailer: mailer.New("", 25, "", "", ""),
	}
}
func newTestServer(t *testing.T, h http.Handler) *httptest.Server{
	t.Helper()
	ts := httptest.NewServer(h)
	return ts

}




func newE2EApplication(t *testing.T, db *sql.DB) *application {
	cfg := config{}
	cfg.limiter.enabled = false
	return &application{
		config: cfg,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		models: data.NewModels(
			data.MovieModel{DB: db},
			data.UserModel{DB: db},
			data.TokenModel{DB: db},
			data.PermissionModel{DB: db},
		),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}
}