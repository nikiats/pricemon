package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"pricemon/internal/web/model"
)

type claims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

func (s *Service) issueToken(userID string) (model.Session, error) {
	now := time.Now()

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
		},
	}).SignedString(s.jwtSecret)
	if err != nil {
		return model.Session{}, err
	}

	return model.Session{
		AccessToken: token,
		ExpiresIn:   int64(s.accessTokenTTL.Seconds()),
	}, nil
}

func (s *Service) ParseToken(token string) (string, error) {
	var c claims

	_, err := jwt.ParseWithClaims(token, &c, func(t *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return "", err
	}

	return c.UserID, nil
}
