package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/WaronLimsakul/Chirpy/internal/auth"
	"github.com/WaronLimsakul/Chirpy/internal/database"
	"github.com/google/uuid"
)

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
		ID:          newUser.ID,
		CreatedAt:   newUser.CreatedAt,
		UpdatedAt:   newUser.UpdatedAt,
		Email:       newUser.Email,
		IsChirpyRed: newUser.IsChirpyRed,
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
		IsChirpyRed:  user.IsChirpyRed,
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

// update user data with new email and password
// need an access token in header
func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	reqToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	userID, err := auth.ValidateJWT(reqToken, cfg.tokenSecret)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	type reqBodyStruct struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	var req reqBodyStruct
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		log.Printf("error decoding at updateUser: %s", err)
		w.WriteHeader(500)
		return
	}

	reqHashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("error hashing password at updateUser: %s", err)
		w.WriteHeader(500)
		return
	}

	updateUserParams := database.UpdateUserEmailPasswordParams{
		Email:          req.Email,
		HashedPassword: reqHashedPassword,
		ID:             userID,
	}
	updatedUser, err := cfg.dbQueries.UpdateUserEmailPassword(r.Context(), updateUserParams)
	if err != nil {
		log.Printf("error updating user data: %s", err)
		w.WriteHeader(500)
		return
	}

	res := User{
		ID:          updatedUser.ID,
		CreatedAt:   updatedUser.CreatedAt,
		UpdatedAt:   updatedUser.UpdatedAt,
		Email:       updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
	}

	resData, err := json.Marshal(res)
	if err != nil {
		log.Printf("error marshalling response at updateUser: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write(resData)
	return
}

// we need
// 1. "Authorization" = "ApiKey <key>" in the header
// 2."event" = "user.upgraded"  in body
// 3."data" = {"user_id" : "..."} in body
func (cfg *apiConfig) reddenUser(w http.ResponseWriter, r *http.Request) {
	reqAPIKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("error at reddenUser: %s", err)
		w.WriteHeader(401)
		return
	}

	if reqAPIKey != cfg.polkaKey {
		w.WriteHeader(401)
		return
	}

	type reqBodyStruct struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	reqBody := reqBodyStruct{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&reqBody)
	if err != nil {
		log.Printf("error decoding body in reddenUser: %s", err)
		w.WriteHeader(500)
		return
	}

	if reqBody.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	userUUID, err := uuid.Parse(reqBody.Data.UserID)
	if err != nil {
		log.Printf("error parsing user uuid at ReddenUser: %s", err)
		w.WriteHeader(500)
		return
	}

	err = cfg.dbQueries.ReddenUserByID(r.Context(), userUUID)
	if err != nil {
		log.Printf("error parsing user uuid at ReddenUser: %s", err)
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(204)
	return
}
