package service

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
	"warehouse_oa/internal/models"
	"warehouse_oa/utils"
)

func Login(username, password string) (map[string]interface{}, error) {
	if username == "" || password == "" {
		return nil, errors.New("nickname or password is empty")
	}

	user, err := CheckPassword(username, password)
	if err != nil {
		return nil, err
	}

	jwtUser := utils.NewJWT()
	token, err := jwtUser.CreateToken(utils.CustomClaims{
		Id:   user.ID,
		Name: user.Nickname,
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),
			ExpiresAt: time.Now().AddDate(1, 0, 0).Unix(),
			Issuer:    "jia_hua",
		},
	})

	user.Password = ""
	return map[string]interface{}{
		"token": token,
		"user":  user,
	}, nil
}

func Register(user *models.User) (map[string]interface{}, error) {
	if user.Username == "" || user.Nickname == "" || user.Password == "" {
		return nil, errors.New("user data is empty")
	}

	user, err := SaveUser(user)
	if err != nil {
		return nil, err
	}

	jwtUser := utils.NewJWT()
	token, err := jwtUser.CreateToken(utils.CustomClaims{
		Id:   user.ID,
		Name: user.Nickname,
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),
			ExpiresAt: time.Now().AddDate(1, 0, 0).Unix(),
			Issuer:    "jia_hua",
		},
	})

	user.Password = ""
	return map[string]interface{}{
		"token": token,
		"user":  user,
	}, nil
}
