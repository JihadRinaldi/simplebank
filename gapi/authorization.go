package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/JihadRinaldi/simplebank/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
)

func (server *Server) AuthorizeUser(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	auth := md.Get(authorizationHeader)
	if len(auth) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}

	authHeader := auth[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	authType := strings.ToLower(fields[0])

	if authType != authorizationBearer {
		return nil, fmt.Errorf("unsupported authorization type: %s", fields[0])
	}

	accessToken := fields[1]

	payload, err := server.TokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return payload, nil
}
