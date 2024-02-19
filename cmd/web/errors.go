package main

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
)

type errorResponse struct {
	Status  int
	Message string
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message string, err error) {
	response := &errorResponse{
		Status:  status,
		Message: message,
	}
	switch status {
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		app.errorLog.Printf("%d:%s\n", status, message)
	default:
		trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
		app.errorLog.Output(2, trace)
	}
	data := app.newTemplateData(r)
	data.Form = response
	app.render(w, status, "error.tmpl", data)
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorLog.Println(err)
	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message, err)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message, errors.New("not found"))
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message, errors.New("method not allowed"))
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, "", err)
}
