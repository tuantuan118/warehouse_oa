package utils

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
	"warehouse_oa/internal/global"
)

type JWT struct {
	SigningKey []byte
}

type CustomClaims struct {
	Id   int
	Name string
	jwt.StandardClaims
}

var (
	TokenExpired     = errors.New("token is expired")
	TokenNotValidYet = errors.New("token not active yet")
	TokenMalformed   = errors.New("token malformed")
	TokenInvalid     = errors.New("token invalid")
)

func NewJWT() *JWT {
	return &JWT{
		SigningKey: []byte(global.ServerConfig.JWTInfo.SigningKey),
	}
}

func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return j.SigningKey, nil
		})
	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			return claims, nil
		}
		return nil, TokenInvalid
	} else {
		return nil, TokenInvalid
	}
}

func (j *JWT) RefreshToken(tokenString string) (string, error) {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(0, 0)
	}
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return j.SigningKey, nil
		})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		jwt.TimeFunc = time.Now
		claims.StandardClaims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
		return j.CreateToken(*claims)
	}

	return "", TokenInvalid
}
