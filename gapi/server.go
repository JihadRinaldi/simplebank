package gapi

import (
	db "github.com/JihadRinaldi/simplebank/db/sqlc"
	pb "github.com/JihadRinaldi/simplebank/pb"
	"github.com/JihadRinaldi/simplebank/token"
	"github.com/JihadRinaldi/simplebank/util"
	"github.com/JihadRinaldi/simplebank/worker"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	Store           db.Store
	TokenMaker      token.Maker
	Config          util.Config
	TaskDistributor worker.TaskDistributor
}

// gRPC server
func NewServer(store db.Store, config util.Config, taskDistributor worker.TaskDistributor) (*Server, error) {
	token, err := token.NewJWTMaker(config.SymmetricKey)
	if err != nil {
		return nil, err
	}

	server := &Server{
		Store:           store,
		TokenMaker:      token,
		Config:          config,
		TaskDistributor: taskDistributor,
	}

	return server, nil
}
