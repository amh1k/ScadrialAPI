package main

import (
	"errors"
	"net/http"
	"time"

	"scadrialapi.abdulmoiz.net/internal/data"
	"scadrialapi.abdulmoiz.net/internal/validator"
)

// @Summary Create authentication token
// @Description Authenticates a user and returns a bearer token.
// @Tags authentication
// @Accept json
// @Produce json
// @Param input body createAuthenticationTokenRequest true "Login credentials"
// @Success 201 {object} authenticationTokenResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 422 {object} validationErrorResponse
// @Failure 500 {object} errorResponse
// @Router /v1/tokens/authentication [post]
func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}
	token, err := app.models.Tokens.New(user.ID, 4*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
