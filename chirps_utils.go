package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/WaronLimsakul/Chirpy/internal/auth"
	"github.com/WaronLimsakul/Chirpy/internal/database"
	"github.com/google/uuid"
)

// 1. parse request body
// 2. gimmick, check 140 characters long or error
// 3. replace kerfuffle/sharbert/fornax with **
// 4. if anything went wrong, respond with json "error" : ??
// func validateChirp(w http.ResponseWriter, r *http.Request) {
// 	// the fields must be exportable or it will not be parsed
// 	// if you don't provide `json:"key"` tag, it will assume the field name as key
// 	type reqStruct struct {
// 		Body string `json:"body"`
// 	}
//
// 	type resStruct struct {
// 		Valid       bool   `json:"valid"`
// 		Error       string `json:"error"`
// 		CleanedBody string `json:"cleaned_body"`
// 	}
//
// 	decoder := json.NewDecoder(r.Body)
// 	req := reqStruct{}
// 	if err := decoder.Decode(&req); err != nil {
// 		log.Printf("error decoding request: %s", err)
// 		w.WriteHeader(500)
// 		errorRes := resStruct{Error: "something went wrong"}
// 		errData, _ := json.Marshal(errorRes) // assume that marshalling will go wel
// 		w.Write(errData)
// 		return
// 	}
//
// 	if len(req.Body) > 140 {
// 		w.WriteHeader(400)
// 		errorRes := resStruct{Error: "Chirp is too long"}
// 		errData, _ := json.Marshal(errorRes) // assume that marshalling will go well
// 		w.Write(errData)
// 		return
// 	}
//
// 	res := resStruct{Valid: true}
// 	resBodyWords := strings.Fields(req.Body)
// 	for i, word := range resBodyWords {
// 		// 3. replace kerfuffle/sharbert/fornax with **
// 		lower := strings.ToLower(word)
// 		if lower == "kerfuffle" {
// 			resBodyWords[i] = "****"
// 		} else if lower == "sharbert" {
// 			resBodyWords[i] = "****"
// 		} else if lower == "fornax" {
// 			resBodyWords[i] = "****"
// 		}
// 	}
//
// 	res.CleanedBody = strings.Join(resBodyWords, " ")
// 	resData, err := json.Marshal(res)
// 	if err != nil {
// 		w.WriteHeader(500)
// 		errorRes := resStruct{Error: "error marshalling response"}
// 		errData, _ := json.Marshal(errorRes) // assume that marshalling will go wel
// 		w.Write(errData)
// 		return
// 	}
// 	w.WriteHeader(200)
// 	w.Write(resData)
// }

// 0. validate user by token in header
// 1. validate chirp (get the above func's logic)
// 2. create chrip in db
// 3. return new chirp in json form
func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(400)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(401)
		return
	}

	type reqBodyStruct struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	req := reqBodyStruct{}
	if err := decoder.Decode(&req); err != nil {
		log.Printf("error decoding request: %s", err)
		w.WriteHeader(500)
		return
	}

	if len(req.Body) > 140 {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("chirp is too long"))
		return
	}

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

	cleanedBody := strings.Join(resBodyWords, " ")

	params := database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userID,
	}

	newChirp, err := cfg.dbQueries.CreateChirp(r.Context(), params)

	res := Chirp{
		ID:        newChirp.ID,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
		Body:      newChirp.Body,
		UserID:    newChirp.UserID,
	}

	resData, err := json.Marshal(res)
	if err != nil {
		log.Println("error marshalling response body")
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(201)
	w.Write(resData)
}

func (cfg *apiConfig) getAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(500)
		return
	}

	var resChirps []Chirp
	for _, chirp := range chirps {
		resChirp := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		resChirps = append(resChirps, resChirp)
	}

	resData, err := json.Marshal(resChirps)
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write(resData)
}

func (cfg *apiConfig) getChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirp_id"))
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(500)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	resChirp := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	resData, err := json.Marshal(resChirp)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write(resData)
	return
}

// check token in the header => get a user id
// use userid +
func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	reqToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("error geting token in deleteChirp: %s", err)
		w.WriteHeader(401)
		return
	}

	userID, err := auth.ValidateJWT(reqToken, cfg.tokenSecret)

	chirpID := r.PathValue("chirp_id")
	if chirpID == "" {
		w.WriteHeader(400)
		return
	}

	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpUUID)
	if chirp.UserID != userID {
		w.WriteHeader(403)
		return
	}

	err = cfg.dbQueries.DeleteChirpByID(r.Context(), chirpUUID)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(204)
	return
}
