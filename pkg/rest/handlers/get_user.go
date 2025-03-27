package handlers

import "net/http"

//nolint:errcheck
func GetUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("List of users"))
}
