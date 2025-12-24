package gapi

import (
	"context"
	"database/sql"
	"testing"

	db "github.com/JihadRinaldi/simplebank/db/sqlc"
	mockdb "github.com/JihadRinaldi/simplebank/mocks"
	pb "github.com/JihadRinaldi/simplebank/pb"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLoginUserAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		req           *pb.LoginUserRequest
		buildStubs    func(store *mockdb.Store)
		checkResponse func(t *testing.T, res *pb.LoginUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: password,
			},
			buildStubs: func(store *mockdb.Store) {
				store.On("GetUser", matchContext(), user.Username).Return(user, nil).Once()
				store.On("CreateSession", matchContext(), mock.Anything).Return(db.Session{}, nil).Once()
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, user.Username, res.User.Username)
				require.Equal(t, user.FullName, res.User.FullName)
				require.Equal(t, user.Email, res.User.Email)
				require.NotEmpty(t, res.AccessToken)
				require.NotEmpty(t, res.RefreshToken)
				require.NotEmpty(t, res.SessionId)
				require.NotNil(t, res.AccessTokenExpiresAt)
				require.NotNil(t, res.RefreshTokenExpiresAt)
			},
		},
		{
			name: "UserNotFound",
			req: &pb.LoginUserRequest{
				Username: "notfound",
				Password: password,
			},
			buildStubs: func(store *mockdb.Store) {
				store.On("GetUser", matchContext(), "notfound").Return(db.User{}, sql.ErrNoRows).Once()
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "IncorrectPassword",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: "wrongpassword",
			},
			buildStubs: func(store *mockdb.Store) {
				store.On("GetUser", matchContext(), user.Username).Return(user, nil).Once()
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "InvalidUsername",
			req: &pb.LoginUserRequest{
				Username: "invalid-user#1",
				Password: password,
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected due to validation failure
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "TooShortPassword",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: "123",
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected due to validation failure
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InternalErrorCreateSession",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: password,
			},
			buildStubs: func(store *mockdb.Store) {
				store.On("GetUser", matchContext(), user.Username).Return(user, nil).Once()
				store.On("CreateSession", matchContext(), mock.Anything).Return(db.Session{}, sql.ErrConnDone).Once()
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := mockdb.NewStore(t)
			taskDistributor := newMockTaskDistributor(t)

			tc.buildStubs(store)

			server := newTestServer(t, store, taskDistributor)

			res, err := server.LoginUser(context.Background(), tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
