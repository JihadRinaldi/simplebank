package api_test

import (
	"os"
	"testing"
	"time"

	"github.com/JihadRinaldi/simplebank/api"
	db "github.com/JihadRinaldi/simplebank/db/sqlc"
	"github.com/JihadRinaldi/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)

	os.Exit(m.Run())
}

func NewTestServer(t *testing.T, store db.Store) *api.Server {
	config := util.Config{
		SymmetricKey:        util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := api.NewServer(store, config)
	require.NoError(t, err)

	return server
}
