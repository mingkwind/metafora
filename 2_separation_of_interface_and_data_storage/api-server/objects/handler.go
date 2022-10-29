package objects

import "net/http"

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		get(w, r)
	case http.MethodPut:
		put(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
