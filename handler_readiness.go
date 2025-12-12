package main

import "net/http"

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{
		"status": "ok",
	}

	respondWithJson(w, 200, resp)
}
