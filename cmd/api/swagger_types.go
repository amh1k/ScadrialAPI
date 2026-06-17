package main

import (
	"time"
)

type healthcheckSystemInfo struct {
	Environment string `json:"environment" example:"development"`
	Version     string `json:"version" example:"1.0.0"`
}

type healthcheckResponse struct {
	Status     string                `json:"status" example:"available"`
	SystemInfo healthcheckSystemInfo `json:"system_info"`
}

type createMovieRequest struct {
	Title   string   `json:"title" example:"The Final Empire"`
	Year    int32    `json:"year" example:"2006"`
	Runtime string   `json:"runtime" example:"126 mins"`
	Genres  []string `json:"genres" example:"fantasy,adventure"`
}

type updateMovieRequest struct {
	Title   *string  `json:"title,omitempty" example:"The Well of Ascension"`
	Year    *int32   `json:"year,omitempty" example:"2007"`
	Runtime *string  `json:"runtime,omitempty" example:"130 mins"`
	Genres  []string `json:"genres,omitempty" example:"fantasy,epic"`
}

type movieSchema struct {
	ID      int64    `json:"id" example:"1"`
	Title   string   `json:"title" example:"The Final Empire"`
	Year    int32    `json:"year" example:"2006"`
	Runtime string   `json:"runtime" example:"126 mins"`
	Genres  []string `json:"genres" example:"fantasy,adventure"`
	Version int32    `json:"version" example:"1"`
}

type metadataSchema struct {
	CurrentPage  int `json:"current_page" example:"1"`
	PageSize     int `json:"page_size" example:"20"`
	FirstPage    int `json:"first_page" example:"1"`
	LastPage     int `json:"last_page" example:"3"`
	TotalRecords int `json:"total_records" example:"45"`
}

type movieResponse struct {
	Movie movieSchema `json:"movie"`
}

type moviesResponse struct {
	Movies   []movieSchema  `json:"movies"`
	Metadata metadataSchema `json:"metadata"`
}

type registerUserRequest struct {
	Name     string `json:"name" example:"Vin"`
	Email    string `json:"email" example:"vin@example.com"`
	Password string `json:"password" example:"supersecret123"`
}

type activateUserRequest struct {
	Token string `json:"token" example:"ABCDEFGHIJKLMNOPQRSTUVWX12"`
}

type userSchema struct {
	ID        int64     `json:"id" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2026-06-17T12:00:00Z"`
	Name      string    `json:"name" example:"Vin"`
	Email     string    `json:"email" example:"vin@example.com"`
	Activated bool      `json:"activated" example:"false"`
}

type userResponse struct {
	User userSchema `json:"user"`
}

type createAuthenticationTokenRequest struct {
	Email    string `json:"email" example:"vin@example.com"`
	Password string `json:"password" example:"supersecret123"`
}

type tokenSchema struct {
	Token  string    `json:"token" example:"ABCDEFGHIJKLMNOPQRSTUVWX12"`
	Expiry time.Time `json:"expiry" example:"2026-06-20T15:04:05Z"`
}

type authenticationTokenResponse struct {
	AuthenticationToken tokenSchema `json:"authentication_token"`
}

type messageResponse struct {
	Message string `json:"message" example:"movie successfully deleted"`
}

type errorResponse struct {
	Error string `json:"error" example:"the requested resource could not be found"`
}

type validationErrorResponse struct {
	Error map[string]string `json:"error"`
}

type authErrorResponse struct {
	Error string `json:"error" example:"invalid or missing authentication token"`
}

type conflictErrorResponse struct {
	Error string `json:"error" example:"unable to update the record due to an edit conflict, please try again"`
}
