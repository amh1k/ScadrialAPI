package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"scadrialapi.abdulmoiz.net/internal/data"
	"scadrialapi.abdulmoiz.net/internal/validator"
)

// @Summary Create movie
// @Description Creates a new movie record.
// @Tags movies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body createMovieRequest true "Movie payload"
// @Success 201 {object} movieResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} authErrorResponse
// @Failure 403 {object} errorResponse
// @Failure 422 {object} validationErrorResponse
// @Failure 500 {object} errorResponse
// @Router /v1/movies [post]
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// @Summary Get movie
// @Description Fetches a single movie by ID.
// @Tags movies
// @Produce json
// @Security BearerAuth
// @Param id path int true "Movie ID"
// @Success 200 {object} movieResponse
// @Failure 401 {object} authErrorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /v1/movies/{id} [get]
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {

		http.NotFound(w, r)
	}
	// fmt.Println("VOLAA")
	// fmt.Println("id is", id)
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// err = app.writeJSON(w, http.StatusOK, movie, nil)
	// if err != nil {
	//     app.logger.Error(err.Error())
	//     http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	// }
}

// @Summary Update movie
// @Description Partially updates a movie by ID.
// @Tags movies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Movie ID"
// @Param X-Expected-Version header string false "Expected current movie version for optimistic locking"
// @Param input body updateMovieRequest true "Movie fields to update"
// @Success 200 {object} movieResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} authErrorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 409 {object} conflictErrorResponse
// @Failure 422 {object} validationErrorResponse
// @Failure 500 {object} errorResponse
// @Router /v1/movies/{id} [patch]
func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return

	}
	movie, err := app.models.Movies.Get(id)

	if err != nil {

		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:

			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.Itoa(int(movie.Version)) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}
	err = app.readJSON(w, r, &input)
	if err != nil {

		app.badRequestResponse(w, r, err)
		return

	}
	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres

	}

	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:

			app.serverErrorResponse(w, r, err)
		}
		return

	}
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// @Summary Delete movie
// @Description Deletes a movie by ID.
// @Tags movies
// @Produce json
// @Security BearerAuth
// @Param id path int true "Movie ID"
// @Success 200 {object} messageResponse
// @Failure 401 {object} authErrorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /v1/movies/{id} [delete]
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// @Summary List movies
// @Description Returns a paginated list of movies with optional search and sorting filters.
// @Tags movies
// @Produce json
// @Security BearerAuth
// @Param title query string false "Movie title full-text search"
// @Param genres query string false "Comma-separated genres"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Param sort query string false "Sort field: id, title, year, runtime, -id, -title, -year, -runtime" default(id)
// @Success 200 {object} moviesResponse
// @Failure 401 {object} authErrorResponse
// @Failure 403 {object} errorResponse
// @Failure 422 {object} validationErrorResponse
// @Failure 500 {object} errorResponse
// @Router /v1/movies [get]
func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}
	v := validator.New()
	qs := r.URL.Query()
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}
	if data.ValidateFilters(v, input.Filters); !v.Valid() {

		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	movies, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {

		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	// fmt.Fprintf(w, "%+v\n", input)

}
