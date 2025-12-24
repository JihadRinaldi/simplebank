package gapi

import (
	"context"
	"os"
	"testing"
	"time"

	db "github.com/JihadRinaldi/simplebank/db/sqlc"
	"github.com/JihadRinaldi/simplebank/util"
	"github.com/JihadRinaldi/simplebank/worker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	config := util.Config{
		SymmetricKey:        util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(store, config, taskDistributor)
	require.NoError(t, err)

	return server
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}

func matchContext() interface{} {
	return mock.MatchedBy(func(ctx interface{}) bool {
		_, ok := ctx.(context.Context)
		return ok
	})
}
