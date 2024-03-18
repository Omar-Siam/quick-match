package models

type UserDetails struct {
	UserID         string `json:"UserID" dynamodbav:"UserID"`
	Email          string `json:"email" dynamodbav:"email"`
	Password       string `json:"password" dynamodbav:"password"`
	PasswordHashed string `json:"password_hashed" dynamodbav:"password_hashed"`
	Name           string `json:"name" dynamodbav:"name"`
	Gender         string `json:"gender" dynamodbav:"gender"`
	Age            int    `json:"age" dynamodbav:"age"`
	Userlocation
}

type Userlocation struct {
	Latitude  float64 `json:"latitude" dynamodbav:"latitude"`
	Longitude float64 `json:"longitude" dynamodbav:"longitude"`
}
