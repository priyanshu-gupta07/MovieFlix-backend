package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// wrap function for multiple middlewares
// func (app *application) wrap(next http.Handler) httprouter.Handle {
// 	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 		// pass httprouter.Params to request context
// 		ctx := context.WithValue(r.Context(), "params", ps)
// 		// call next middleware with new context
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	}
// }

func (app *application) routes() http.Handler {
	router := httprouter.New()

	// Define your routes here
	router.HandlerFunc(http.MethodGet, "/v1/status", app.GetStatus)
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.getAllMovies)

	// Add more routes as needed

	return app.enableCORS(router)
}
