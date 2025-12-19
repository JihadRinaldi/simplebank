package util_test

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/JihadRinaldi/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestPassword(t *testing.T) {
	password := util.RandomString(6)

	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	err = util.CheckPassword(password, hashedPassword)
	require.NoError(t, err)

	wrongPassword := util.RandomString(7)
	err = util.CheckPassword(wrongPassword, hashedPassword)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

}
