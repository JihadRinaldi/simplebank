package gapi

import (
	"context"
	"database/sql"
	"testing"

	db "github.com/JihadRinaldi/simplebank/db/sqlc"
	mockdb "github.com/JihadRinaldi/simplebank/mocks"
	pb "github.com/JihadRinaldi/simplebank/pb"
	"github.com/JihadRinaldi/simplebank/util"
	"github.com/JihadRinaldi/simplebank/worker"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		req           *pb.CreateUserRequest
		buildStubs    func(store *mockdb.Store, taskDistributor *mockTaskDistributor)
		checkResponse func(t *testing.T, res *pb.CreateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.Store, taskDistributor *mockTaskDistributor) {
				arg := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email:    user.Email,
					},
				}

				store.On("CreateUserTx", matchContext(), mock.MatchedBy(func(params db.CreateUserTxParams) bool {
					// Check all params except HashedPassword and AfterCreate callback
					err := util.CheckPassword(password, params.HashedPassword)
					return params.Username == arg.Username &&
						params.FullName == arg.FullName &&
						params.Email == arg.Email &&
						err == nil &&
						params.AfterCreate != nil
				})).Return(db.CreateUserTxResult{
					User: user,
				}, nil).Once()

				taskDistributor.On("DistributeTaskSendVerifyEmail", matchContext(), mock.MatchedBy(func(payload *worker.PayloadSendVerifyEmail) bool {
					return payload.Username == user.Username
				}), mock.Anything).Return(nil).Once()
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				createdUser := res.GetUser()
				require.Equal(t, user.Username, createdUser.Username)
				require.Equal(t, user.FullName, createdUser.FullName)
				require.Equal(t, user.Email, createdUser.Email)
			},
		},
		{
			name: "InternalError",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.Store, taskDistributor *mockTaskDistributor) {
				store.On("CreateUserTx", matchContext(), mock.Anything).Return(db.CreateUserTxResult{}, sql.ErrConnDone).Once()
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "InvalidUsername",
			req: &pb.CreateUserRequest{
				Username: "invalid-user#1",
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.Store, taskDistributor *mockTaskDistributor) {
				// No DB call expected due to validation failure
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidEmail",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    "invalid-email",
			},
			buildStubs: func(store *mockdb.Store, taskDistributor *mockTaskDistributor) {
				// No DB call expected due to validation failure
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "TooShortPassword",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: "123",
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.Store, taskDistributor *mockTaskDistributor) {
				// No DB call expected due to validation failure
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidFullName",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: "123",
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.Store, taskDistributor *mockTaskDistributor) {
				// No DB call expected due to validation failure
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := mockdb.NewStore(t)
			taskDistributor := newMockTaskDistributor(t)

			tc.buildStubs(store, taskDistributor)

			server := newTestServer(t, store, taskDistributor)

			res, err := server.CreateUser(context.Background(), tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}

// mockTaskDistributor is a mock for worker.TaskDistributor
type mockTaskDistributor struct {
	mock.Mock
}

func newMockTaskDistributor(t *testing.T) *mockTaskDistributor {
	return &mockTaskDistributor{}
}

func (m *mockTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *worker.PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	args := m.Called(ctx, payload, opts)
	return args.Error(0)
}
