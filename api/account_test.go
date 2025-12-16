package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JihadRinaldi/simplebank/api"
	db "github.com/JihadRinaldi/simplebank/db/sqlc"
	mockdb "github.com/JihadRinaldi/simplebank/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateAccountAPI(t *testing.T) {
	account := db.Account{
		ID:       1,
		Owner:    "testuser",
		Balance:  0,
		Currency: "USD",
	}

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.Store)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			buildStubs: func(store *mockdb.Store) {
				arg := db.CreateAccountRequest{
					Owner:    account.Owner,
					Balance:  0,
					Currency: account.Currency,
				}
				store.On("CreateAccount", matchContext(), arg).Return(account, nil).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"owner":    account.Owner,
				"currency": account.Currency,
			},
			buildStubs: func(store *mockdb.Store) {
				arg := db.CreateAccountRequest{
					Owner:    account.Owner,
					Balance:  0,
					Currency: account.Currency,
				}
				store.On("CreateAccount", matchContext(), arg).Return(db.Account{}, context.DeadlineExceeded).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidCurrency",
			body: gin.H{
				"owner":    account.Owner,
				"currency": "INVALID",
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidOwner",
			body: gin.H{
				"owner":    "",
				"currency": account.Currency,
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "MissingCurrency",
			body: gin.H{
				"owner": account.Owner,
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "MissingOwner",
			body: gin.H{
				"currency": account.Currency,
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

			server := api.NewServer(store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/accounts"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetAccountAPI(t *testing.T) {
	account := db.Account{
		ID:       1,
		Owner:    "testuser",
		Balance:  0,
		Currency: "USD",
	}

	testCases := []struct {
		name          string
		accountID     string
		buildStubs    func(store *mockdb.Store)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: "1",
			buildStubs: func(store *mockdb.Store) {
				store.On("GetAccount", matchContext(), int64(1)).Return(account, nil).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: "2",
			buildStubs: func(store *mockdb.Store) {
				store.On("GetAccount", matchContext(), int64(2)).Return(db.Account{}, sql.ErrNoRows).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: "1",
			buildStubs: func(store *mockdb.Store) {
				store.On("GetAccount", matchContext(), int64(1)).Return(db.Account{}, context.DeadlineExceeded).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: "abc",
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "ZeroID",
			accountID: "0",
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

			server := api.NewServer(store)
			recorder := httptest.NewRecorder()

			url := "/accounts/" + tc.accountID
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccountAPI(t *testing.T) {
	accounts := []db.Account{
		{ID: 1, Owner: "user1", Balance: 0, Currency: "USD"},
		{ID: 2, Owner: "user2", Balance: 0, Currency: "USD"},
	}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mockdb.Store)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "page=1&size=5",
			buildStubs: func(store *mockdb.Store) {
				arg := db.ListAccountsParams{
					Limit:  5,
					Offset: 0,
				}
				store.On("ListAccounts", matchContext(), arg).Return(accounts, nil).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},
		{
			name:  "InternalError",
			query: "page=1&size=5",
			buildStubs: func(store *mockdb.Store) {
				arg := db.ListAccountsParams{
					Limit:  5,
					Offset: 0,
				}
				store.On("ListAccounts", matchContext(), arg).Return(nil, context.DeadlineExceeded).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "InvalidPage",
			query: "page=0&size=5",
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "InvalidSizeTooSmall",
			query: "page=1&size=4",
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "InvalidSizeTooLarge",
			query: "page=1&size=11",
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "MissingPage",
			query: "size=5",
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "MissingSize",
			query: "page=1",
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

			server := api.NewServer(store)
			recorder := httptest.NewRecorder()

			url := "/accounts"
			if tc.query != "" {
				url += "?" + tc.query
			}
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

// helper

func matchContext() interface{} {
	return mock.MatchedBy(func(ctx interface{}) bool {
		_, ok := ctx.(context.Context)
		return ok
	})
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []db.Account
	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accounts, gotAccounts)
}
