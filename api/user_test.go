package api_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/JihadRinaldi/simplebank/db/sqlc"
	mockdb "github.com/JihadRinaldi/simplebank/mocks"
	"github.com/JihadRinaldi/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.Store)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.Store) {
				store.On("CreateUser", matchContext(), mock.MatchedBy(func(arg db.CreateUserParams) bool {
					err := util.CheckPassword(password, arg.HashedPassword)
					return arg.Username == user.Username &&
						arg.FullName == user.FullName &&
						arg.Email == user.Email &&
						err == nil
				})).Return(user, nil).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username":  user.Username,
				"password":  "password",
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.Store) {
				store.On("CreateUser", matchContext(), mock.Anything).Return(db.User{}, sql.ErrConnDone).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		}, {
			name: "InvalidPasswordSize",
			body: gin.H{
				"username":  user.Username,
				"password":  "Wrong",
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := mockdb.NewStore(t)
			tc.buildStubs(store)

			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			request.Header.Set("Content-Type", "application/json")

			server.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateUserAPI_DuplicateUsername(t *testing.T) {
	user, password := randomUser(t)

	store := mockdb.NewStore(t)

	store.On("CreateUser", matchContext(), mock.MatchedBy(func(arg db.CreateUserParams) bool {
		return arg.Username == user.Username
	})).Return(user, nil).Once()

	store.On("CreateUser", matchContext(), mock.MatchedBy(func(arg db.CreateUserParams) bool {
		return arg.Username == user.Username
	})).Return(db.User{}, &pq.Error{Code: "23505"}).Once()

	server := NewTestServer(t, store)

	body := gin.H{
		"username":  user.Username,
		"password":  password,
		"full_name": user.FullName,
		"email":     user.Email,
	}
	data, err := json.Marshal(body)
	require.NoError(t, err)

	recorder1 := httptest.NewRecorder()
	request1, err := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(data))
	require.NoError(t, err)
	request1.Header.Set("Content-Type", "application/json")

	server.ServeHTTP(recorder1, request1)
	require.Equal(t, http.StatusOK, recorder1.Code)

	recorder2 := httptest.NewRecorder()
	request2, err := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(data))
	require.NoError(t, err)
	request2.Header.Set("Content-Type", "application/json")

	server.ServeHTTP(recorder2, request2)
	require.Equal(t, http.StatusInternalServerError, recorder2.Code)
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		require.NoError(t, err)
	}

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}
