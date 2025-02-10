package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"rwa/internal/dto"
)

var (
	errCreateUser    error = errors.New("Create user error. %s")
	errLoginUser     error = errors.New("Login user error. %s")
	errRequestBody   error = errors.New("Unmarshal body error. %s")
	errHTTPMethod    error = errors.New("Bad HTTP method. %s")
	errCreateSession error = errors.New("Create session error. %s")
	errGetUser       error = errors.New("Get user error. %s")
	errMarshalResp   error = errors.New("Marshal response error. %s")
	errUserUpd       error = errors.New("Update user error. %s")
)

func writeError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	errors := dto.Error{dto.Errors{[]string{fmt.Sprintf("Handle error: %s", err.Error())}}}
	res, err := json.Marshal(errors)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	w.Write(res)
}
