package models

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
)

//Token JWT claims struct
type Token struct {
	UserID uint
	jwt.StandardClaims
}

//Account struct to user account
type Account struct {
	gorm.Model
	Email string `json:"email"`
	Password string `json:"password"`
	Token string `json:"token":sql:"-"'`
}

//Create create account for Admin!
func (account *Account)Create()(map[string]interface{}){

}

//Validate checking for an account in the database
func (account *Account)Validate()(map[string]interface{},bool){
	temp := &Account{}
	
}


