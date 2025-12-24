package gapi

import (
	"context"
	"database/sql"
	"testing"

	db "github.com/JihadRinaldi/simplebank/db/sqlc"
	mockdb "github.com/JihadRinaldi/simplebank/mocks"
	pb "github.com/JihadRinaldi/simplebank/pb"
	"github.com/JihadRinaldi/simplebank/util"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestVerifyEmailAPI(t *testing.T) {
	user, _ := randomUser(t)
	emailId := util.RandomInt(1, 1000)
	secretCode := util.RandomString(32)

	testCases := []struct {
		name          string
		req           *pb.VerifyEmailRequest
		buildStubs    func(store *mockdb.Store)
		checkResponse func(t *testing.T, res *pb.VerifyEmailResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.VerifyEmailRequest{
				EmailId:    emailId,
				SecretCode: secretCode,
			},
			buildStubs: func(store *mockdb.Store) {
				arg := db.VerifyEmailTxParams{
					EmailId:    emailId,
					SecretCode: secretCode,
				}

				verifiedUser := db.User{
					Username:          user.Username,
					HashedPassword:    user.HashedPassword,
					FullName:          user.FullName,
					Email:             user.Email,
					PasswordChangedAt: user.PasswordChangedAt,
					CreatedAt:         user.CreatedAt,
					IsEmailVerified:   true,
				}

				txResult := db.VerifyEmailTxResult{
					User: verifiedUser,
					VerifyEmail: db.VerifyEmail{
						ID:         emailId,
						Username:   user.Username,
						Email:      user.Email,
						SecretCode: secretCode,
					},
				}

				store.On("VerifyEmailTx", matchContext(), mock.MatchedBy(func(params db.VerifyEmailTxParams) bool {
					return params.EmailId == arg.EmailId &&
						params.SecretCode == arg.SecretCode
				})).Return(txResult, nil).Once()
			},
			checkResponse: func(t *testing.T, res *pb.VerifyEmailResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.True(t, res.IsVerified)
			},
		},
		{
			name: "InvalidEmailId",
			req: &pb.VerifyEmailRequest{
				EmailId:    0,
				SecretCode: secretCode,
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected due to validation failure
			},
			checkResponse: func(t *testing.T, res *pb.VerifyEmailResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidSecretCode",
			req: &pb.VerifyEmailRequest{
				EmailId:    emailId,
				SecretCode: "",
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected due to validation failure
			},
			checkResponse: func(t *testing.T, res *pb.VerifyEmailResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "NotFound",
			req: &pb.VerifyEmailRequest{
				EmailId:    emailId,
				SecretCode: secretCode,
			},
			buildStubs: func(store *mockdb.Store) {
				store.On("VerifyEmailTx", matchContext(), mock.Anything).Return(db.VerifyEmailTxResult{}, sql.ErrNoRows).Once()
			},
			checkResponse: func(t *testing.T, res *pb.VerifyEmailResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "InternalError",
			req: &pb.VerifyEmailRequest{
				EmailId:    emailId,
				SecretCode: secretCode,
			},
			buildStubs: func(store *mockdb.Store) {
				store.On("VerifyEmailTx", matchContext(), mock.Anything).Return(db.VerifyEmailTxResult{}, sql.ErrConnDone).Once()
			},
			checkResponse: func(t *testing.T, res *pb.VerifyEmailResponse, err error) {
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

			res, err := server.VerifyEmail(context.Background(), tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
