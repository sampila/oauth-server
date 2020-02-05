package rest

import (
	"github.com/sampila/go-utils/rest_errors"
	"github.com/sampila/oauth-server/domain/user"
	"github.com/mercadolibre/golang-restclient/rest"
	"time"
	"encoding/json"
	"errors"
	"log"
)

var (
	usersRestClient = rest.RequestBuilder{
		BaseURL: "http://localhost:8084",
		Timeout: 100 * time.Millisecond,
	}
)

type RestUsersRepository interface {
	LoginUser(string, string) (*user.User, rest_errors.RestErr)
}

type usersRepository struct{}

func NewRestUsersRepository() RestUsersRepository {
	return &usersRepository{}
}

func (r *usersRepository) LoginUser(email string, password string) (*user.User, rest_errors.RestErr) {
	log.Println(email)
	log.Println(password)
	request := user.UserLoginRequest{
		Email:    email,
		Password: password,
	}

	response := usersRestClient.Post("/login", request)

	if response == nil || response.Response == nil {
		return nil, rest_errors.NewInternalServerError("invalid restclient response when trying to login user", errors.New("restclient error"))
	}

	if response.StatusCode > 299 {
		apiErr, err := rest_errors.NewRestErrorFromBytes(response.Bytes())
		if err != nil {
			return nil, rest_errors.NewInternalServerError("invalid error interface when trying to login user", err)
		}
		return nil, apiErr
	}

	var usr user.User
	if err := json.Unmarshal(response.Bytes(), &usr); err != nil {
		return nil, rest_errors.NewInternalServerError("error when trying to unmarshal users login response", errors.New("json parsing error"))
	}
	return &usr, nil
}
