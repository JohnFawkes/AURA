package config

import (
	"net/http"
)

func Reload(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement reload logic
	w.WriteHeader(http.StatusOK)
}
