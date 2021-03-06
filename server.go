package main

import (
	"encoding/json"
	"log"
	"strconv"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	mysql "github.com/imrenagi/go-oauth2-mysql"
	"github.com/jmoiron/sqlx"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"github.com/sampila/oauth-server/repository/rest"
)

type service struct {
	restUsersRepo rest.RestUsersRepository
}

var (
	userRepo = rest.NewRestUsersRepository()
)

func main() {
	db, err := sqlx.Connect("mysql", "root:aptikma@tcp(127.0.0.1:3306)/oauth_db?parseTime=true")
	if err != nil {
		log.Fatalln(err)
	}

	manager := manage.NewDefaultManager()

	clientStore, _ := mysql.NewClientStore(db)
	//clientStore = store.NewClientStore()
	clientStore.Create(&models.Client{
		ID:     "111111",
		Secret: "supersecret",
		Domain: "http://localhost:8084",
		UserID: "12",
	})
	clientStore.Create(&models.Client{
		ID:     "3sdGzJ7rKkyZjPU15SWEqEK5Uwho9NDp",
		Secret: "9UFhraag61zgv01AJtVeDaxivoGLYhBb",
		Domain: "http://localhost:8084",
	})
	tokenStore, _ := mysql.NewTokenStore(db)
	manager.MapClientStorage(clientStore)
	manager.MapTokenStorage(tokenStore)

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)
	//auth password granty type handler
	srv.SetPasswordAuthorizationHandler(func(username, password string) (userID string, err error) {
		//To-Do check to api login
		// Authenticate the user against the Users API:
		s := &service{
			restUsersRepo : userRepo,
		}
		respond, restErr := s.restUsersRepo.LoginUser(username, password)
		if restErr == nil {
			resData := respond["data"].(map[string]interface{})
			userID = strconv.Itoa(int(resData["ID"].(float64)))
		}
		return
	})

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		err := srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		srv.HandleTokenRequest(w, r)
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		token, err := srv.ValidationBearerToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data := map[string]interface{}{
			"expires_in": int64(token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()).Sub(time.Now()).Seconds()),
			"client_id":  token.GetClientID(),
			"user_id":    token.GetUserID(),
		}
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		e.Encode(data)
	})

	log.Println("oauth server running on port 9096")
	log.Fatal(http.ListenAndServe(":9096", nil))
}
