package api

import (
	"errors"
	"fmt"
	"net/http"

	db "github.com/JihadRinaldi/simplebank/db/sqlc"
	"github.com/JihadRinaldi/simplebank/token"
	"github.com/gin-gonic/gin"
)

type CreateTransferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req CreateTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if valid := server.transferValidation(ctx, req.FromAccountID, req.ToAccountID, req); !valid {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := server.Store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (server *Server) transferValidation(ctx *gin.Context, fromAccountID int64, toAccountID int64, arg CreateTransferRequest) bool {
	fromAccount, err := server.Store.GetAccount(ctx, fromAccountID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	toAccount, err := server.Store.GetAccount(ctx, toAccountID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	authPayload := ctx.MustGet(AuthorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("unauthorized access")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return false
	}

	if fromAccount.Balance < arg.Amount {
		err := fmt.Errorf("account [%d] has insufficient balance: %d", fromAccount.ID, fromAccount.Balance)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}

	if fromAccount.Currency != arg.Currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", fromAccount.ID, fromAccount.Currency, arg.Currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}

	if toAccount.Currency != arg.Currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", toAccount.ID, toAccount.Currency, arg.Currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}

	return true
}
