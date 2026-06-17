//go:build e2e
// +build e2e

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"scadrialapi.abdulmoiz.net/internal/data"
)

func createAuthenticatedUser(t *testing.T, testApp *application, permissions ...string) string {
	user := &data.User{
		Name:      "Test User",
		Email:     "test@example.com",
		Activated: true,
	}
	if err := user.Password.Set("12345678"); err != nil {
		t.Fatal(err)
	}
	if err := testApp.models.Users.Insert(user); err != nil {
		t.Fatal(err)
	}

	token, err := testApp.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		t.Fatal(err)
	}
	if len(permissions) > 0 {
		if err := testApp.models.Permissions.AddForUser(user.ID, permissions...); err != nil {
			t.Fatal(err)
		}
	}
	return token.Plaintext
}

func createTestMovie(t *testing.T, testApp *application) *data.Movie {
	movie := &data.Movie{
		Title:   "The Two Towers",
		Year:    2002,
		Runtime: 179,
		Genres:  []string{"fantasy", "adventure"},
	}
	if err := testApp.models.Movies.Insert(movie); err != nil {
		t.Fatal(err)
	}
	return movie
}

func TestMovieShow(t *testing.T) {
	testDB := newTestDB(t)
	testApp := newE2EApplication(t, testDB)
	ts := newTestServer(t, testApp.routes())
	token := createAuthenticatedUser(t, testApp, "movies:read")
	movie := createTestMovie(t, testApp)
	// fmt.Println(token)
	// fmt.Println(movie.ID)
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v1/movies/%d", ts.URL, movie.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got status %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var out struct {
		Movie data.Movie `json:"movie"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Movie.Title != movie.Title {
		t.Errorf("got title %q, want %q", out.Movie.Title, movie.Title)
	}
}
func TestMovieUpdate(t *testing.T) {
	testDB := newTestDB(t)
	testApp := newE2EApplication(t, testDB)
	ts := newTestServer(t, testApp.routes())
	token := createAuthenticatedUser(t, testApp, "movies:write")
	movie := createTestMovie(t, testApp)

	updateBody := `{"title":"The Two Towers (Extended Edition)"}`
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/v1/movies/%d", ts.URL, movie.ID), strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got status %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var out struct {
		Movie data.Movie `json:"movie"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Movie.Title != "The Two Towers (Extended Edition)" {
		t.Errorf("got title %q, want updated title", out.Movie.Title)
	}
	if out.Movie.Version != movie.Version+1 {
		t.Errorf("got version %d, want %d", out.Movie.Version, movie.Version+1)
	}
}
func TestMovieDelete(t *testing.T) {
	testDB := newTestDB(t)
	testApp := newE2EApplication(t, testDB)
	ts := newTestServer(t, testApp.routes())

	token := createAuthenticatedUser(t, testApp, "movies:write", "movies:read")
	movie := createTestMovie(t, testApp)

	delReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/v1/movies/%d", ts.URL, movie.ID), nil)
	delReq.Header.Set("Authorization", "Bearer "+token)

	delResp, err := ts.Client().Do(delReq)
	if err != nil {
		t.Fatal(err)
	}
	defer delResp.Body.Close()

	if delResp.StatusCode != http.StatusOK {
		t.Fatalf("got status %d, want %d", delResp.StatusCode, http.StatusOK)
	}

	getReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v1/movies/%d", ts.URL, movie.ID), nil)
	getReq.Header.Set("Authorization", "Bearer "+token)

	getResp, err := ts.Client().Do(getReq)
	if err != nil {
		t.Fatal(err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected movie to be gone, got status %d", getResp.StatusCode)
	}
}
func TestMovieList(t *testing.T) {
	testDB := newTestDB(t)
	testApp := newE2EApplication(t, testDB)
	ts := newTestServer(t, testApp.routes())

	token := createAuthenticatedUser(t, testApp, "movies:read")

	titles := []string{"The Fellowship of the Ring", "The Two Towers", "The Return of the King"}
	for _, title := range titles {
		m := &data.Movie{Title: title, Year: 2001, Runtime: 178, Genres: []string{"fantasy"}}
		if err := testApp.models.Movies.Insert(m); err != nil {
			t.Fatal(err)
		}
	}
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/v1/movies?page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got status %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var out struct {
		Movies   []data.Movie  `json:"movies"`
		Metadata data.Metadata `json:"metadata"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.Movies) != len(titles) {
		t.Errorf("got %d movies, want %d", len(out.Movies), len(titles))
	}
}
