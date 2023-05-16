package model

import "github.com/dgrijalva/jwt-go"

type CustomJwtClaims struct {
	Iss  string `json:"iss"`
	Dest string `json:"dest"`
	Aud  string `json:"aud"`
	Sub  string `json:"sub"`
	Exp  int    `json:"exp"`
	Nbf  int    `json:"nbf"`
	Iat  int    `json:"iat"`
	Jti  string `json:"jti"`
	Sid  string `json:"sid"`
	jwt.StandardClaims
}
