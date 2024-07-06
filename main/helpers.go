package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// readJSON reads json from request body into data. We only accept a single json value in the body
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1048576 // max one megabyte in request body
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	// we only allow one entry in the json file
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only have a single JSON value")
	}

	return nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, wrap ...string) error {
	var js []byte
	var err error
	if len(wrap) > 0 {

		wrapper := make(map[string]interface{})
		wrapper[wrap[0]] = data

		js, err = json.Marshal(wrapper)
		if err != nil {
			return err
		}
	} else {
		js, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (app *application)errorJSON(w http.ResponseWriter, err error,status ...int) {
	statusCode := http.StatusBadRequest

	if(len(status)>0) {
		statusCode = status[0]
	}
	type jsonError struct {
		Message string `json:"message"`
	}
	theError := jsonError{
		Message: err.Error(),
	}
	app.writeJSON(w, statusCode, theError, "error")
}
