package objects

import "net/http"

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		GetObject(w, r)
	case "PUT":
		putObject(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
