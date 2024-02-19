package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) router() http.Handler {
	//TODO:Implement internal server responses
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	fileServer := http.FileServer(http.Dir("../../ui/static/"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	return router
}
