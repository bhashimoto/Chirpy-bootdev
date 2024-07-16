package main

type User struct {
	ID int `json:"id"`
	Email string `json:"email"`
	IsChirpyRed bool `json:"is_chirpy_red"`
}

type UserCredential struct {
	User
	Password []byte `json:"password"`

}
