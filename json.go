package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type JSON struct {
	Streams		[]struct{
		Width		int		`json:"width"`
		Height		int		`json:"height"`
	}`json:"streams"`
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func unMarshal(data []byte) (JSON, error) {
	var info JSON
	if err := json.Unmarshal(data, &info); err != nil {
		return JSON{}, err
	}
	return info, nil
}
