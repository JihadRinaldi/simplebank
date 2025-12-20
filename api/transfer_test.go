package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JihadRinaldi/simplebank/api"
	db "github.com/JihadRinaldi/simplebank/db/sqlc"
	mockdb "github.com/JihadRinaldi/simplebank/mocks"
	"github.com/JihadRinaldi/simplebank/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCreateTransferAPI(t *testing.T) {
	amount := int64(10)

	fromAccount := db.Account{
		ID:       1,
		Owner:    "user1",
		Balance:  100,
		Currency: "USD",
	}

	toAccount := db.Account{
		ID:       2,
		Owner:    "user2",
		Balance:  100,
		Currency: "USD",
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.Store)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.Store) {
				// Mock GetAccount for fromAccount
				store.On("GetAccount", matchContext(), fromAccount.ID).Return(fromAccount, nil).Once()
				// Mock GetAccount for toAccount
				store.On("GetAccount", matchContext(), toAccount.ID).Return(toAccount, nil).Once()

				// Mock TransferTx
				arg := db.TransferTxParams{
					FromAccountID: fromAccount.ID,
					ToAccountID:   toAccount.ID,
					Amount:        amount,
				}
				result := db.TransferTxResult{
					Transfer: db.Transfer{
						ID:            1,
						FromAccountID: fromAccount.ID,
						ToAccountID:   toAccount.ID,
						Amount:        amount,
					},
					FromAccount: db.Account{
						ID:       fromAccount.ID,
						Owner:    fromAccount.Owner,
						Balance:  fromAccount.Balance - amount,
						Currency: fromAccount.Currency,
					},
					ToAccount: db.Account{
						ID:       toAccount.ID,
						Owner:    toAccount.Owner,
						Balance:  toAccount.Balance + amount,
						Currency: toAccount.Currency,
					},
					FromEntry: db.Entry{
						ID:        1,
						AccountID: fromAccount.ID,
						Amount:    -amount,
					},
					ToEntry: db.Entry{
						ID:        2,
						AccountID: toAccount.ID,
						Amount:    amount,
					},
				}
				store.On("TransferTx", matchContext(), arg).Return(result, nil).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransferResult(t, recorder.Body)
			},
		},
		{
			name: "InsufficientBalance",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          1000, // More than balance
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.Store) {
				store.On("GetAccount", matchContext(), fromAccount.ID).Return(fromAccount, nil).Once()
				store.On("GetAccount", matchContext(), toAccount.ID).Return(toAccount, nil).Once()
				// TransferTx should not be called
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "FromAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "EUR",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.Store) {
				// Your code fetches both accounts before checking currency
				store.On("GetAccount", matchContext(), fromAccount.ID).Return(fromAccount, nil).Once()
				store.On("GetAccount", matchContext(), toAccount.ID).Return(toAccount, nil).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ToAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "EUR",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.Store) {
				eurFromAccount := fromAccount
				eurFromAccount.Currency = "EUR"
				store.On("GetAccount", matchContext(), fromAccount.ID).Return(eurFromAccount, nil).Once()
				store.On("GetAccount", matchContext(), toAccount.ID).Return(toAccount, nil).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "FromAccountNotFound",
			body: gin.H{
				"from_account_id": 999,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.Store) {
				store.On("GetAccount", matchContext(), int64(999)).Return(db.Account{}, context.DeadlineExceeded).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "ToAccountNotFound",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   999,
				"amount":          amount,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.Store) {
				store.On("GetAccount", matchContext(), fromAccount.ID).Return(fromAccount, nil).Once()
				store.On("GetAccount", matchContext(), int64(999)).Return(db.Account{}, context.DeadlineExceeded).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "TransferTxInternalError",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.Store) {
				store.On("GetAccount", matchContext(), fromAccount.ID).Return(fromAccount, nil).Once()
				store.On("GetAccount", matchContext(), toAccount.ID).Return(toAccount, nil).Once()

				arg := db.TransferTxParams{
					FromAccountID: fromAccount.ID,
					ToAccountID:   toAccount.ID,
					Amount:        amount,
				}
				store.On("TransferTx", matchContext(), arg).Return(db.TransferTxResult{}, context.DeadlineExceeded).Once()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidFromAccountID",
			body: gin.H{
				"from_account_id": 0,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidToAccountID",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   0,
				"amount":          amount,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidAmount",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          0,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NegativeAmount",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          -10,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
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
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "MissingFromAccountID",
			body: gin.H{
				"to_account_id": toAccount.ID,
				"amount":        amount,
				"currency":      "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "MissingToAccountID",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.Store) {
				// No DB call expected
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "MissingAmount",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"currency":        "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, api.AuthorizationTypeBearer, fromAccount.Owner, time.Minute)
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

			url := "/transfers"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.TokenMaker)
			server.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

// Helper function
func requireBodyMatchTransferResult(t *testing.T, body *bytes.Buffer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotResult db.TransferTxResult
	err = json.Unmarshal(data, &gotResult)
	require.NoError(t, err)

	// Verify the structure is valid
	require.NotEmpty(t, gotResult.Transfer)
	require.NotEmpty(t, gotResult.FromAccount)
	require.NotEmpty(t, gotResult.ToAccount)
	require.NotEmpty(t, gotResult.FromEntry)
	require.NotEmpty(t, gotResult.ToEntry)
}
