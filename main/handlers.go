package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/julienschmidt/httprouter"
	"github.com/priyanshu-gupta07/MovieFlix-backend/models"
	"github.com/priyanshu-gupta07/MovieFlix-backend/validator"
	"golang.org/x/crypto/bcrypt"
)

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

// get all movies /req;
func (app *application) getAllMovies(w http.ResponseWriter, r *http.Request) {
	//get all movies from db
	movies, err := app.models.Db.GetAllMovies()
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

// get all genres /req;
func (app *application) GetAllGenres(w http.ResponseWriter, r *http.Request) {
	//get all genres from db
	genres, err := app.models.Db.GetAllGenres()
	if err != nil {
		app.logger.Println(err)
		app.errorJSON(w, err)
		return
	}
	//write genres to response
	err = app.writeJSON(w, http.StatusOK, genres, "genres")
	if err != nil {
		app.logger.Println(err)
		app.errorJSON(w, err)
		return
	}
}

func (app *application) GetLatestMovies(w http.ResponseWriter, r *http.Request) {
	//get latest featured movies on the platform
	movies, err := app.models.Db.GetLatestMovies()
	if err != nil {
		app.logger.Println(err)
		app.errorJSON(w, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, movies, "movies")

	if err != nil {
		app.logger.Println(err)
		app.errorJSON(w, err)
		return
	}
}

func (app *application) getAllMoviesByGenre(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	genreID, err := strconv.Atoi(params.ByName("genre_id"))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	movies, err := app.models.Db.GetMoviesByGenre(genreID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.logger.Println(genreID)

	err = app.writeJSON(w, http.StatusOK, movies, "movies")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}

func (app *application) getOneMovie(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.errorJSON(w, errors.New("invalid id parameter"))
		return
	}

	movie, err := app.models.Db.GetMovie(id)
	if err != nil {
		app.errorJSON(w, errors.New("failed to fetch the movie"))
		return
	}

	err = app.writeJSON(w, http.StatusOK, movie, "movie")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}

// login user
type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// custom claims
type CustomClaims struct {
	UserName string `json:"name"`
	UserType string `json:"user_type"`
	jwt.StandardClaims
}

func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		app.badRequest(w, r, errors.New("invalid request body"))
		return
	}

	// get user from the database
	user, err := app.models.Db.GetUserByEmail(creds.Email)
	if err != nil {
		app.badRequest(w, r, errors.New("invalid email or password"))
		return
	}

	// compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password))
	if err != nil {
		app.badRequest(w, r, errors.New("invalid email or password"))
		return
	}

	// custom claims
	claims := CustomClaims{
		UserType: user.UserType,
		UserName: user.FullName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "movieapp",
			Subject:   strconv.Itoa(user.ID),
			NotBefore: time.Now().Unix(),
			Audience:  "movieapp",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Sign the token with the secret key
	signedToken, err := token.SignedString([]byte(app.config.jwt.secret))
	if err != nil {
		app.errorJSON(w, errors.New("can't generate jwt token"), http.StatusInternalServerError)
		return
	}

	var resp struct {
		OK      bool   `json:"ok"`
		Token   string `json:"token"`
		Message string `json:"message"`
	}

	resp.OK = true
	resp.Token = signedToken
	resp.Message = "user is successfully logged in!"

	err = app.writeJSON(w, http.StatusOK, resp)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}

// signUp a new user
func (app *application) signUp(w http.ResponseWriter, r *http.Request) {
	var payload models.User

	// read json from the body
	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.badRequest(w, r, errors.New("invalid json request"))
		return
	}

	v := validator.New()

	// check email is valid or not
	v.IsEmail(payload.Email, "email", "invalid email address")

	// check email is already exist or not if email is not exit then continue otherwise return error
	u, _ := app.models.Db.GetUserByEmail(payload.Email)
	if u != nil {
		v.AddError("email", "email is already exits")
	}

	// check user password is valid or not
	v.IsValidPassword(payload.Password, "password")

	// check your full name is valid or not.
	v.Required(payload.FullName, "full_name", "Full Name is required")
	v.IsLength(payload.FullName, "full_name", 5, 55)
	v.IsValidFullName(payload.FullName, "full_name")

	if !v.Valid() {
		err := app.writeJSON(w, http.StatusBadRequest, v)
		if err != nil {
			app.badRequest(w, r, err)
			return
		}
		return
	}

	// convert the password into hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)

	if err != nil {
		app.errorJSON(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}

	// insert new user into the database
	err = app.models.Db.InsertUser(payload.FullName, payload.Email, string(hashedPassword))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// send the response
	var resp struct {
		OK      bool
		Message string
	}

	// return ok response with message
	resp.OK = true
	resp.Message = "User have sign up successfully now login with your credentials!"
	app.writeJSON(w, http.StatusOK, resp)
}
