package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"scadrialapi.abdulmoiz.net/internal/assert"
	"scadrialapi.abdulmoiz.net/internal/data"
	"scadrialapi.abdulmoiz.net/internal/data/mocks"
)
var(
	tokenTest = "A7kP9mX2qR8tY4nL6wC3vB1dEf"
)


func TestRateLimit(t *testing.T) {
	app := &application{}
	app.config.limiter.enabled = true
	app.config.limiter.rps = 2
	app.config.limiter.burst = 4
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := app.rateLimit(next)
	doReq := func(ip string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip + ":1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr.Code

	}
	t.Run("requests within burst succeed", func(t *testing.T) {
		for i := 0; i < 4; i++ {
			if code := doReq("1.2.3.4"); code != http.StatusOK {
				t.Errorf("request %d: got %d, want %d", i, code, http.StatusOK)
			}
		}
	})

	t.Run("request exceeding burst is limited", func(t *testing.T) {
		if code := doReq("1.2.3.4"); code != http.StatusTooManyRequests {
			t.Errorf("got %d, want %d", code, http.StatusTooManyRequests)
		}
	})

	t.Run("different IP is tracked independently", func(t *testing.T) {
		if code := doReq("5.6.7.8"); code != http.StatusOK {
			t.Errorf("got %d, want %d", code, http.StatusOK)
		}
	})
}

func TestAuthenticate(t *testing.T) {
	mockUserModel := &mocks.UserModel{}
	var gotUser *data.User
	testApp := NewTestApplication(t, mockUserModel)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = testApp.contextGetUser(r)
		w.WriteHeader(http.StatusOK)
	})
	handler := testApp.authenticate(next)
	authFunc := func() int{
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("Authorization", "Bearer A7kP9mX2qR8tY4nL6wC3vB1dEf")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr.Code
	}
	code := authFunc()
	assert.Equal(t, code, http.StatusOK)
	if gotUser == nil || gotUser == data.AnonymousUser {
		t.Fatalf("expected an authenticated user in context, got %#v", gotUser)
	}

}