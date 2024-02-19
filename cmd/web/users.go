package main

import (
	"errors"
	"net/http"
	"reservista/internal/data"
	"reservista/internal/validator"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgtype"
)

const SecretKey = "secret"

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
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", d)
		return
	}
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			form.Validator.AddError("email", "a user with this email already exists.")
			d.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", d)
		case errors.Is(err, data.ErrDuplicateNickname):
			form.Validator.AddError("nickname", "a user with this nickname already exists.")
			d.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", d)
		default:
			app.serverErrorResponse(w, r, err)

		}
		return
	}
	//TODO: perform permissions for user

	//TODO: create activation token for user
	//TODO: write activate user endpoint

	d.User = user
	app.render(w, http.StatusOK, "index.tmpl", d)
}

func (app *application) createUserHandlerGet(w http.ResponseWriter, r *http.Request) {
	d := app.newTemplateData(r)
	d.Form = userCreateForm{}
	app.render(w, http.StatusOK, "signup.tmpl", d)
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {

}

func (app *application) LogIn(w http.ResponseWriter, r *http.Request) {
	//retrieve user's credentials
	var form userCreateForm
	d := app.newTemplateData(r)
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	var user *data.User
	//find it in db
	user, err = app.models.Users.GetByNickname(form.Nickname)
	if err != nil {
		if errors.Is(err, data.ErrNoRecord) {
			app.badRequestResponse(w, r, err)
			return
		}
	}
	//compare passwords
	ok, err := user.Password.Matches(form.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !ok {
		form.Validator.AddError("nickname", "wrong nickname or password.")
		form.Validator.AddError("password", "wrong nickname or password.")
		d.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signin.tmpl", d)
	}

	id := user.ID.Bytes
	//OR
	//TODO: use encodeText
	//create claims
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    string(id[:]),
		ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
	})
	//create token
	token, err := claims.SignedString([]byte(SecretKey))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	//create cookie
	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

func (app *application) LogOut(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	token, err := jwt.ParseWithClaims(cookie.Value, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})
	if err != nil {
		//TODO: return unauthorized access error
		app.serverErrorResponse(w, r, err)
		return
	}
	claims := token.Claims.(*jwt.StandardClaims)
	d := app.newTemplateData(r)
	id := pgtype.UUID{}
	err = id.DecodeText(nil, []byte(claims.Issuer))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	user, err := app.models.Users.GetByID(id)
	if err != nil {
		//TODO: handle possible panic, because access was gained, but user is not found
		app.serverErrorResponse(w, r, err)
		return
	}
	d.User = user
	app.render(w, http.StatusOK, "index.tmpl", d)
}

//TODO: activate User
//TODO: log in (save token and compare it after, middleware)
//TODO: log out
//TODO: profile overview, posts, comments and user settings
