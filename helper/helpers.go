package helper

import (
	"fmt"
	"net/http"

	"github.com/nu7hatch/gouuid"

	"github.com/SpringerPE/graphite-du-report/logging"
)

func ErrorResponse(w http.ResponseWriter, msg string, err error) {
	logging.LogError(msg, err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, msg)
}

func GenerateSecret() (string, error) {
	uuid, err := uuid.NewV4()
	secret := uuid.String()
	return secret, err
}
