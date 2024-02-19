package main

import (
	"errors"
	"net/http"
	"reservista/internal/data"
	"reservista/internal/validator"
)

type userCreateForm struct {
	Name                string `form:"name"`
	Nickname            string `form:"nickname"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) testModel() error {
	//user := data.User{
	//	Name:      "buzuk",
	//	Nickname:  "buzuk",
	//	Email:     "buzuk@gmail.com",
	//	Password:  []byte("buz"),
	//	Activated: true,
	//}
	user, err := app.models.Users.GetByNickname("buzuk")
	if err != nil {
		return err
	}
	user.Email = "ExampleEmail@yahoo.com"
	err = app.models.Users.Update(user)
	app.infoLog.Println(user)
	return nil
}

func (app *application) createUserHandlerPost(w http.ResponseWriter, r *http.Request) {
	var form userCreateForm
	d := app.newTemplateData(r)
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	user := &data.User{
		Name:      form.Name,
		Nickname:  form.Nickname,
		Email:     form.Email,
		Activated: false,
	}
	err = user.Password.Set(form.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	form.Validator = *(validator.New())
	if data.ValidateUser(&(form.Validator), user); !form.Validator.Valid() {
		d.Form = form
		app.render(w, http.StatusUnprocessableEntity, "register.tmpl", d)
		return
	}
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			form.Validator.AddError("email", "a user with this email already exists.")
			d.Form = form
			app.render(w, http.StatusUnprocessableEntity, "register.tmpl", d)
		case errors.Is(err, data.ErrDuplicateNickname):
			form.Validator.AddError("nickname", "a user with this nickname already exists.")
			d.Form = form
			app.render(w, http.StatusUnprocessableEntity, "register.tmpl", d)
		default:
			app.serverErrorResponse(w, r, err)

		}
		return
	}
	//TODO: perform permissions for user

	//TODO: create activation token for user
	//TODO: write activate user endpoint

	d.User = user
	app.render(w, http.StatusOK, "home.tmpl", d)
}

func (app *application) createUserHandlerGet(w http.ResponseWriter, r *http.Request) {
	d := app.newTemplateData(r)
	d.Form = userCreateForm{}
	app.render(w, http.StatusOK, "register.tmpl", d)
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {

}

//TODO: activate User
//TODO: log in (save token and compare it after, middleware)
//TODO: log out
//TODO: profile overview, posts, comments and user settings
