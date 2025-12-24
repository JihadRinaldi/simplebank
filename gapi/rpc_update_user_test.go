package gapi

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	db "github.com/JihadRinaldi/simplebank/db/sqlc"
	mockdb "github.com/JihadRinaldi/simplebank/mocks"
	pb "github.com/JihadRinaldi/simplebank/pb"
	"github.com/JihadRinaldi/simplebank/token"
	"github.com/JihadRinaldi/simplebank/util"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUpdateUserAPI(t *testing.T) {
	user, _ := randomUser(t)

	newName := util.RandomOwner()
	newEmail := util.RandomEmail()
	newPassword := util.RandomString(6)
	invalidEmail := "invalid-email"

	testCases := []struct {
		name          string
		req           *pb.UpdateUserRequest
		buildStubs    func(store *mockdb.Store)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.UpdateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.Store) {
				arg := db.UpdateUserParams{
					Username: user.Username,
					FullName: sql.NullString{String: newName, Valid: true},
					Email:    sql.NullString{String: newEmail, Valid: true},
				}

				updatedUser := db.User{
					Username:          user.Username,
					HashedPassword:    user.HashedPassword,
					FullName:          newName,
					Email:             newEmail,
					PasswordChangedAt: user.PasswordChangedAt,
					CreatedAt:         user.CreatedAt,
					IsEmailVerified:   user.IsEmailVerified,
				}

				store.On("UpdateUser", matchContext(), mock.MatchedBy(func(params db.UpdateUserParams) bool {
					return params.Username == arg.Username &&
						params.FullName.String == arg.FullName.String &&
						params.FullName.Valid == arg.FullName.Valid &&
						params.Email.String == arg.Email.String &&
						params.Email.Valid == arg.Email.Valid &&
						!params.HashedPassword.Valid
				})).Return(updatedUser, nil).Once()
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, user.Username, res.User.Username)
				require.Equal(t, newName, res.User.FullName)
				require.Equal(t, newEmail, res.User.Email)
			},
		},
		{
			name: "OnlyFullName",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
			},
			buildStubs: func(store *mockdb.Store) {
				arg := db.UpdateUserParams{
					Username: user.Username,
					FullName: sql.NullString{String: newName, Valid: true},
				}

				updatedUser := db.User{
					Username:          user.Username,
					HashedPassword:    user.HashedPassword,
					FullName:          newName,
					Email:             user.Email,
					PasswordChangedAt: user.PasswordChangedAt,
					CreatedAt:         user.CreatedAt,
					IsEmailVerified:   user.IsEmailVerified,
				}

				store.On("UpdateUser", matchContext(), mock.MatchedBy(func(params db.UpdateUserParams) bool {
					return params.Username == arg.Username &&
						params.FullName.String == arg.FullName.String &&
						params.FullName.Valid == arg.FullName.Valid &&
						!params.Email.Valid &&
						!params.HashedPassword.Valid
				})).Return(updatedUser, nil).Once()
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, user.Username, res.User.Username)
				require.Equal(t, newName, res.User.FullName)
			},
		},
		{
			name: "OnlyEmail",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.Store) {
				arg := db.UpdateUserParams{
					Username: user.Username,
					Email:    sql.NullString{String: newEmail, Valid: true},
				}

				updatedUser := db.User{
					Username:          user.Username,
					HashedPassword:    user.HashedPassword,
					FullName:          user.FullName,
					Email:             newEmail,
					PasswordChangedAt: user.PasswordChangedAt,
					CreatedAt:         user.CreatedAt,
					IsEmailVerified:   user.IsEmailVerified,
				}

				store.On("UpdateUser", matchContext(), mock.MatchedBy(func(params db.UpdateUserParams) bool {
					return params.Username == arg.Username &&
						!params.FullName.Valid &&
						params.Email.String == arg.Email.String &&
						params.Email.Valid == arg.Email.Valid &&
						!params.HashedPassword.Valid
				})).Return(updatedUser, nil).Once()
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, user.Username, res.User.Username)
				require.Equal(t, newEmail, res.User.Email)
			},
		},
		{
			name: "OnlyPassword",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				Password: &newPassword,
			},
			buildStubs: func(store *mockdb.Store) {
				updatedUser := db.User{
					Username:          user.Username,
					HashedPassword:    user.HashedPassword,
					FullName:          user.FullName,
					Email:             user.Email,
					PasswordChangedAt: user.PasswordChangedAt,
					CreatedAt:         user.CreatedAt,
					IsEmailVerified:   user.IsEmailVerified,
				}

				store.On("UpdateUser", matchContext(), mock.MatchedBy(func(params db.UpdateUserParams) bool {
					err := util.CheckPassword(newPassword, params.HashedPassword.String)
					return params.Username == user.Username &&
						!params.FullName.Valid &&
						!params.Email.Valid &&
						params.HashedPassword.Valid &&
						params.PasswordChangedAt.Valid &&
						err == nil
				})).Return(updatedUser, nil).Once()
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, user.Username, res.User.Username)
			},
		},
		{
			name: "InvalidEmail",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				Email:    &invalidEmail,
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected due to validation failure
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "ExpiredToken",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected due to expired token
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "NoAuthorization",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected due to missing authorization
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "UpdateOtherUser",
			req: &pb.UpdateUserRequest{
				Username: "other_user",
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected due to permission denied
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := mockdb.NewStore(t)
			taskDistributor := newMockTaskDistributor(t)

			tc.buildStubs(store)

			server := newTestServer(t, store, taskDistributor)
			ctx := tc.buildContext(t, server.TokenMaker)

			res, err := server.UpdateUser(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}

func newContextWithBearerToken(t *testing.T, tokenMaker token.Maker, username string, duration time.Duration) context.Context {
	accessToken, _, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)

	bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)
	md := metadata.MD{
		authorizationHeader: []string{
			bearerToken,
		},
	}

	return metadata.NewIncomingContext(context.Background(), md)
}
