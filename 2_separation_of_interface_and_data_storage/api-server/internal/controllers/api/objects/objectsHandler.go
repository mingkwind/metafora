package objects

import (
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetObjects(w, r)
	case http.MethodPut:
		PutObjects(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
