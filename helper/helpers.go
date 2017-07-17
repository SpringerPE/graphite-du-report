package helper

import (
	"fmt"
	"net/http"

	"github.com/SpringerPE/graphite-du-report/logging"
)

func ErrorResponse(w http.ResponseWriter, msg string, err error) {
	logging.LogError(msg, err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, msg)
}
