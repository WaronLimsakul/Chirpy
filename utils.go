package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/WaronLimsakul/Chirpy/internal/auth"
	"github.com/WaronLimsakul/Chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type LoggedInUser struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

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

// create user in db, we need "email" and "password" key in json body
// - not return hashed password in response
func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type reqBodyStruct struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	req := reqBodyStruct{}
	if err := decoder.Decode(&req); err != nil {
		log.Println("error decoding request")
		w.WriteHeader(500)
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(500)
		return
	}

	email := req.Email
	userParams := database.CreateUserParams{
		Email:          email,
		HashedPassword: hashedPassword,
	}

	newUser, err := cfg.dbQueries.CreateUser(r.Context(), userParams)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	taggedNewUser := User{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	}

	resData, err := json.Marshal(taggedNewUser)
	if err != nil {
		log.Println("error marshalling data")
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(201)
	w.Write(resData)
	return
}

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

// Request body has => "password" and "email"
// 0. Decode body
// 1. Set expire duration
// 2. Get user by email
// 3. Compare password with the hash one
// 4. Create access token for user
// 5. Create refresh token for user
// 6. Respond: Invalid password -> 401 , Valid -> 200
func (cfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	// Go guarantee zeo-initialize, so int is 0 if we reqBodyStruct{}
	type reqBodyStruct struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)

	// 0.
	var req reqBodyStruct
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("%s", err)
		return
	}

	// 1.
	expiresIn := time.Hour

	// 2.
	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	// 3.
	unMatch := auth.CheckPasswordHash(req.Password, user.HashedPassword)
	if unMatch != nil {
		w.WriteHeader(401)
		return
	}

	// 4.
	token, err := auth.MakeJWT(user.ID, cfg.tokenSecret, expiresIn)
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(500)
		return
	}

	// 5.
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(500)
		return
	}

	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 1440), // 60 days
	}

	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		log.Printf("%s", err)
		return
	}

	// 6.
	resBody := LoggedInUser{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}

	resData, err := json.Marshal(resBody)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write(resData)
	return
}

// 1. extract token from header
// 2. look up in database
// 3. create a new JWT
func (cfg *apiConfig) refreshUser(w http.ResponseWriter, r *http.Request) {
	reqToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	refreshToken, err := cfg.dbQueries.GetRefreshToken(r.Context(), reqToken)
	// check if
	// 1. token exists
	// 2. exceeds expire date
	// 3. got revoked
	if err != nil || refreshToken.ExpiresAt.Before(time.Now()) || refreshToken.RevokedAt.Valid {
		w.WriteHeader(401)
		return
	}

	newToken, err := auth.MakeJWT(refreshToken.UserID, cfg.tokenSecret, time.Hour)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	type resBody struct {
		Token string `json:"token"`
	}
	res := resBody{
		Token: newToken,
	}

	resData, err := json.Marshal(res)
	w.WriteHeader(200)
	w.Write(resData)
}

// revoke refresh token from the request
func (cfg *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	reqToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(400)
		return
	}

	err = cfg.dbQueries.RevokeToken(r.Context(), reqToken)
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(204)
	return
}
