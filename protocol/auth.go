package protocol

import (
	"errors"
	"net/http"
)

var (
	errAuthenticationFailed = errors.New("authentication failed")
)

type Credential struct {
	Username string `json:"-"`
	Password string `json:"-"`
}

func getCredential(r *http.Request) (Credential, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return Credential{}, errAuthenticationFailed
	}

	return Credential{Username: username, Password: password}, nil
}
