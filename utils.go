package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// 1. parse request body
// 2. gimmick, check 140 characters long or error
// 3. replace kerfuffle/sharbert/fornax with **
// 4. if anything went wrong, respond with json "error" : ??
func validateChirp(w http.ResponseWriter, r *http.Request) {
	// the fields must be exportable or it will not be parsed
	// if you don't provide `json:"key"` tag, it will assume the field name as key
	type reqStruct struct {
		Body string `json:"body"`
	}

	type resStruct struct {
		Valid       bool   `json:"valid"`
		Error       string `json:"error"`
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	req := reqStruct{}
	if err := decoder.Decode(&req); err != nil {
		log.Printf("error decoding request: %s", err)
		w.WriteHeader(500)
		errorRes := resStruct{Error: "something went wrong"}
		errData, _ := json.Marshal(errorRes) // assume that marshalling will go wel
		w.Write(errData)
		return
	}

	if len(req.Body) > 140 {
		w.WriteHeader(400)
		errorRes := resStruct{Error: "Chirp is too long"}
		errData, _ := json.Marshal(errorRes) // assume that marshalling will go well
		w.Write(errData)
		return
	}

	res := resStruct{Valid: true}
	resBodyWords := strings.Fields(req.Body)
	for i, word := range resBodyWords {
		// 3. replace kerfuffle/sharbert/fornax with **
		lower := strings.ToLower(word)
		if lower == "kerfuffle" {
			resBodyWords[i] = "****"
		} else if lower == "sharbert" {
			resBodyWords[i] = "****"
		} else if lower == "fornax" {
			resBodyWords[i] = "****"
		}
	}

	res.CleanedBody = strings.Join(resBodyWords, " ")
	resData, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(500)
		errorRes := resStruct{Error: "error marshalling response"}
		errData, _ := json.Marshal(errorRes) // assume that marshalling will go wel
		w.Write(errData)
		return
	}
	w.WriteHeader(200)
	w.Write(resData)
}
