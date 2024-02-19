package main

import (
	"fmt"
	"net/http"
	"time"
)

func (app *application) serve() error {
	srv := http.Server{
		Addr:         fmt.Sprintf("localhost:%d", app.cfg.port),
		Handler:      app.router(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     app.errorLog,
	}
	app.infoLog.Printf("server listens on port %d\n", app.cfg.port)
	err := srv.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}
