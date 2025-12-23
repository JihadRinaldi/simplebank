package token_test

import (
	"testing"
	"time"

	"github.com/JihadRinaldi/simplebank/token"
	"github.com/JihadRinaldi/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {
	maker, err := token.NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	username := "username"
	duration := util.RandomInt(1, 10)
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(time.Duration(duration) * time.Minute)

	token, payload, err := maker.CreateToken(username, time.Duration(duration)*time.Minute)
	require.NoError(t, err)
	require.NotNil(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotNil(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredJWTToken(t *testing.T) {
	username := "username"
	maker, err := token.NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	jwtToken, payload, err := maker.CreateToken(username, -time.Minute)
	require.NoError(t, err)
	require.NotNil(t, jwtToken)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(jwtToken)
	require.Error(t, err)
	require.EqualError(t, err, token.ErrExpiredToken.Error())
	require.Nil(t, payload)
}
