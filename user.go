package main

type User struct {
	ID int `json:"id"`
	Email string `json:"email"`
}

type UserCredential struct {
	User
	Password []byte `json:"password"`

}
