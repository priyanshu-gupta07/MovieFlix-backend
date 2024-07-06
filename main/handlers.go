package main

import "net/http"

// const (
// 	defaultPage    = 1
// 	defaultPerPage = 3
// )

func (app *application) GetStatus(w http.ResponseWriter, r *http.Request) {
	currentStatus := AppStatus{
		Status:      "Available",
		Environment: app.config.env,
		Version:     version,
	}
	err := app.writeJSON(w, http.StatusOK, currentStatus, "app_status")
	if err != nil {
		app.logger.Println(err)
	}
}

type MoviePayload struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Year        string         `json:"year"`
	ReleaseDate string         `json:"release_date"`
	Runtime     string         `json:"runtime"`
	ImageID     string         `json:"image_id"`
	MovieGenre  map[int]string `json:"genres"`
}

//get all movies /req;
func (app *application) getAllMovies(w http.ResponseWriter, r *http.Request) {
	//get all movies from db
	movies, err := app.models.Db.GetAllMovies("the")
	if err != nil {
		app.logger.Println(err)
		app.errorJSON(w, err)
		return
	}
	//write movies to response
	err = app.writeJSON(w, http.StatusOK, movies, "movies")
	if err != nil {
		app.logger.Println(err)
		app.errorJSON(w, err)
		return
	}
}