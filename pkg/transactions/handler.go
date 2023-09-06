package transactions

import (
	"financial-app/pkg/accounts"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var (
	transactionIDRequired = "transaction id required"
	currencyNotSupported  = "currency is not supported"
	accountsNotSame       = "accounts cannot be the same"
)

type TransactionHandler struct {
	Service Service

	Logger *zap.SugaredLogger
}

// router sets up all the routes for tranfer service
func (h *TransactionHandler) Router(routerGroup *gin.RouterGroup) {
	routerGroup.GET("transactions/:id", h.load)
	routerGroup.GET("transactions", h.loadAll)
	routerGroup.POST("transactions", h.transfer)
	routerGroup.DELETE("transactions/:id", h.clean)
}

// load retrieves a transaction by ID
func (h *TransactionHandler) load(context *gin.Context) {
	id := context.Param("id")
	if id == "" {
		h.Logger.Error("no transaction id found")

		context.JSON(http.StatusBadRequest, gin.H{
			"error": transactionIDRequired,
		})
		return
	}

	transaction, err := h.Service.Load(context, id)
	if err != nil {
		h.Logger.Error(err)

		context.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, transaction)
}

// loadAll retrieves all the completed transactions
func (h *TransactionHandler) loadAll(context *gin.Context) {
	transactions := h.Service.LoadAll(context)

	context.JSON(http.StatusOK, transactions)
}

// transferRequest
type transactionRequest struct {
	SourceAccountID string  `json:"source_account_id" validate:"required,uuid"`
	TargetAccountID string  `json:"target_account_id" validate:"required,uuid"`
	Amount          float64 `json:"amount" validate:"required,numeric,gt=0"`
	Currency        string  `json:"currency" validate:"currency"`
}

func transactionRequestFromTransactionDomain(p transactionRequest) Transaction {
	return Transaction{
		ID:              nextTransactionID(), // Generate a new uuid
		SourceAccountID: p.SourceAccountID,
		TargetAccountID: p.TargetAccountID,
		Amount:          p.Amount,
		Currency:        p.Currency,
	}
}

// validCurrency validates if the given currency is supported
func validCurrency(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		return accounts.IsSupportedCurrency(currency)
	}
	return false
}

// transfer performs a new transaction from source to target account
func (h *TransactionHandler) transfer(context *gin.Context) {
	var transactionReq transactionRequest

	if err := context.ShouldBindJSON(&transactionReq); err != nil {
		h.Logger.Error(err)

		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	v := validator.New()
	v.RegisterValidation("currency", validCurrency)

	err := v.Var(transactionReq.Currency, "currency")
	if err != nil {
		h.Logger.Error(err)

		context.JSON(http.StatusBadRequest, gin.H{
			"error": currencyNotSupported,
		})
		return
	}

	err = v.Struct(transactionReq)
	if err != nil {
		h.Logger.Error(err)

		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = v.VarWithValue(transactionReq.SourceAccountID,
		transactionReq.TargetAccountID, "necsfield")
	if err != nil {
		h.Logger.Error(err)

		context.JSON(http.StatusConflict, gin.H{
			"error": accountsNotSame,
		})
		return
	}

	txn := transactionRequestFromTransactionDomain(transactionReq)
	transaction, err := h.Service.Transfer(context, txn)
	if err != nil {
		noSourceAccountFound := accounts.ErrFetchingAccount(txn.SourceAccountID).Error()
		noTargetAccountFound := accounts.ErrFetchingAccount(txn.TargetAccountID).Error()
		if err.Error() == noSourceAccountFound ||
			err.Error() == noTargetAccountFound {
			context.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Insufficient balance error
		if strings.Contains(err.Error(), "insufficient") {
			context.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
			return
		}

		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, transaction)
}

// clean deletes a transaction by ID
func (h *TransactionHandler) clean(context *gin.Context) {
	id := context.Param("id")
	if id == "" {
		h.Logger.Error("no transaction id found")

		context.JSON(http.StatusBadRequest, gin.H{
			"error": transactionIDRequired,
		})
		return
	}

	err := h.Service.Clean(context, id)
	if err != nil {
		h.Logger.Error(err)

		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		id: "Deleted",
	})
}
