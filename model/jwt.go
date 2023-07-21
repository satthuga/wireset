package model

import (
	"github.com/golang-jwt/jwt/v5"
)

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
	jwt.RegisteredClaims
}
