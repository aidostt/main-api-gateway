package main

import (
	"errors"
	"fmt"
	"github.com/go-playground/form/v4"
	"github.com/jackc/pgtype"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrBadRequest     = errors.New("bad request")
	ErrInternalServer = errors.New("internal server error")
)

func (app *application) decodePostForm(r *http.Request, dst any) error {

	err := r.ParseForm()
	if err != nil {
		return err
	}
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError
		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		return err
	}
	return nil
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear: time.Now().Year(),
	}
}

func (app *application) retrieveID(r *http.Request) (pgtype.UUID, error) {
	params := httprouter.ParamsFromContext(r.Context())
	var id pgtype.UUID
	strID := params.ByName("id")
	err := id.Set(strID)
	if err != nil {
		switch err.Error() {
		case fmt.Sprintf("[]byte must be 16 bytes to convert to UUID: %d", len(strID)), fmt.Sprintf("cannot convert %v to UUID", strID):
			return pgtype.UUID{}, ErrBadRequest
		case fmt.Sprintf("cannot parse UUID %v", strID):
			return pgtype.UUID{}, ErrNotFound
		default:
			return pgtype.UUID{}, ErrInternalServer
		}
	}
	return id, nil
}
