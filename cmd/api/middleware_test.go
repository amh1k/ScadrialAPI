package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"scadrialapi.abdulmoiz.net/internal/assert"
	"scadrialapi.abdulmoiz.net/internal/data"
	"scadrialapi.abdulmoiz.net/internal/data/mocks"
)

var (
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
	mockUserModel := &mocks.UserModel{
		GetForTokenFn: func(scope, token string) (*data.User, error) {
			return &data.User{ID: 1, Activated: true}, nil
		},
	}
	mockPermissionModel := &mocks.PermissionModel{}
	var gotUser *data.User
	testApp := NewTestApplication(t, mockUserModel, mockPermissionModel)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = testApp.contextGetUser(r)
		w.WriteHeader(http.StatusOK)
	})
	handler := testApp.authenticate(next)
	authFunc := func() int {
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

func TestAuthenticateAndPermissionPipeline(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		userMock   *mocks.UserModel
		permMock   *mocks.PermissionModel
		wantStatus int
	}{
		{
			name:       "no authorization header",
			authHeader: "",
			userMock:   &mocks.UserModel{},
			permMock:   &mocks.PermissionModel{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid token, has permission",
			authHeader: "Bearer " + tokenTest,
			userMock: &mocks.UserModel{
				GetForTokenFn: func(scope, token string) (*data.User, error) {
					return &data.User{ID: 1, Activated: true}, nil
				},
			},
			permMock: &mocks.PermissionModel{
				GetAllForUserFn: func(userID int64) (data.Permissions, error) {
					return data.Permissions{"movies:read"}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid token, lacks permission",
			authHeader: "Bearer " + tokenTest,
			userMock: &mocks.UserModel{
				GetForTokenFn: func(scope, token string) (*data.User, error) {
					return &data.User{ID: 1, Activated: true}, nil
				},
			},
			permMock: &mocks.PermissionModel{
				GetAllForUserFn: func(userID int64) (data.Permissions, error) {
					return data.Permissions{}, nil
				},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "valid token, inactive user",
			authHeader: "Bearer " + tokenTest,
			userMock: &mocks.UserModel{
				GetForTokenFn: func(scope, token string) (*data.User, error) {
					return &data.User{ID: 1, Activated: false}, nil
				},
			},
			permMock:   &mocks.PermissionModel{}, // never reached
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "malformed authorization header",
			authHeader: "Token abc",
			userMock:   &mocks.UserModel{},
			permMock:   &mocks.PermissionModel{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "token fails validation (too short)",
			authHeader: "Bearer short",
			userMock:   &mocks.UserModel{},
			permMock:   &mocks.PermissionModel{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "token not found in store",
			authHeader: "Bearer " + tokenTest,
			userMock: &mocks.UserModel{
				GetForTokenFn: func(scope, token string) (*data.User, error) {
					return nil, data.ErrRecordNotFound
				},
			},
			permMock:   &mocks.PermissionModel{},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewTestApplication(t, tt.userMock, tt.permMock)
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			handler := app.authenticate(app.requirePermission("movies:read", next))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			assert.Equal(t, rr.Code, tt.wantStatus)
		})
	}
}
