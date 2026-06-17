package main

import (
	"io"
	"log/slog"
	"testing"

	"scadrialapi.abdulmoiz.net/internal/data"
	"scadrialapi.abdulmoiz.net/internal/mailer"
)
func NewTestApplication(t *testing.T, mockUsers data.UserModelInterface) *application {
	cfg := config{}
	cfg.limiter.enabled = false

	return &application{
		config: cfg,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		models: data.Models{
			Users: mockUsers,
		},
		mailer: mailer.New("", 25, "", "", ""),
	}
}