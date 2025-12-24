package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/JihadRinaldi/simplebank/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hashedPassowrd, err := util.HashPassword("secret")
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassowrd,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)

	user2, err := testQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.WithinDuration(t, user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestUpdateUserFullName(t *testing.T) {
	oldUser := createRandomUser(t)

	newFullName := util.RandomOwner()
	updatedUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		FullName: sql.NullString{
			String: newFullName,
			Valid:  bool(newFullName != ""),
		},
		Username: oldUser.Username,
	})

	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.WithinDuration(t, oldUser.PasswordChangedAt, updatedUser.PasswordChangedAt, time.Second)
	require.WithinDuration(t, oldUser.CreatedAt, updatedUser.CreatedAt, time.Second)
}

func TestUpdateUserEmail(t *testing.T) {
	oldUser := createRandomUser(t)

	newEmail := util.RandomEmail()
	updatedUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Email: sql.NullString{
			String: newEmail,
			Valid:  bool(newEmail != ""),
		},
		Username: oldUser.Username,
	})
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, newEmail, updatedUser.Email)
	require.WithinDuration(t, oldUser.PasswordChangedAt, updatedUser.PasswordChangedAt, time.Second)
	require.WithinDuration(t, oldUser.CreatedAt, updatedUser.CreatedAt, time.Second)
}

func TestUpdateUserPassword(t *testing.T) {
	oldUser := createRandomUser(t)

	newHashedPassword, err := util.HashPassword("newsecret")
	require.NoError(t, err)

	updatedUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		HashedPassword: sql.NullString{
			String: newHashedPassword,
			Valid:  bool(newHashedPassword != ""),
		},
		Username: oldUser.Username,
	})
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.WithinDuration(t, oldUser.CreatedAt, updatedUser.CreatedAt, time.Second)
}

func TestUpdateUserAllFields(t *testing.T) {
	oldUser := createRandomUser(t)

	newHashedPassword, err := util.HashPassword("newsecret")
	require.NoError(t, err)
	newFullName := util.RandomOwner()
	newEmail := util.RandomEmail()

	updatedUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		HashedPassword: sql.NullString{
			String: newHashedPassword,
			Valid:  bool(newHashedPassword != ""),
		},
		FullName: sql.NullString{
			String: newFullName,
			Valid:  bool(newFullName != ""),
		},
		Email: sql.NullString{
			String: newEmail,
			Valid:  bool(newEmail != ""),
		},
		Username: oldUser.Username,
	})
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, newEmail, updatedUser.Email)
	require.WithinDuration(t, oldUser.CreatedAt, updatedUser.CreatedAt, time.Second)
}
