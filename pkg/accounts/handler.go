package accounts

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var (
	accountIDRequired    = "account id required"
	currencyNotSupported = "currency is not supported"
)

type AccountHandler struct {
	Service Service

	Logger *zap.SugaredLogger
}

// Router sets up all the routes for account service
func (h *AccountHandler) Router(routerGroup *gin.RouterGroup) {
	routerGroup.GET("accounts/:id", h.load)
	routerGroup.GET("accounts", h.loadAll)
	routerGroup.POST("accounts", h.register)
	routerGroup.DELETE("accounts/:id", h.clean)
}

// load retrieves an account by ID
func (h *AccountHandler) load(context *gin.Context) {
	id := context.Param("id")
	if id == "" {
		h.Logger.Error("no account id found")

		context.JSON(http.StatusBadRequest, gin.H{
			"error": accountIDRequired,
		})
		return
	}

	acct, err := h.Service.LoadAccount(context, id)
	if err != nil {
		h.Logger.Error(err)

		context.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, acct)
}

// loadAll retrieves all the registered accounts
func (h *AccountHandler) loadAll(context *gin.Context) {
	accounts := h.Service.Accounts(context)

	context.JSON(http.StatusOK, accounts)
}

// validCurrency validates if the given currency is supported
func validCurrency(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		return IsSupportedCurrency(currency)
	}
	return false
}

// storeRequest
type storeRequest struct {
	Balance  float64 `json:"balance" validate:"required"`
	Currency string  `json:"currency" validate:"currency"`
}

func accountRequestFromAccountDomain(p storeRequest) Account {
	return Account{
		ID:       nextAccountID(), // Generate a new uuid
		Balance:  p.Balance,
		Currency: p.Currency,
	}
}

// register registers a new account
func (h *AccountHandler) register(context *gin.Context) {
	var storeReq storeRequest
	if err := context.ShouldBindJSON(&storeReq); err != nil {
		h.Logger.Error(err)

		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	validate := validator.New()
	validate.RegisterValidation("currency", validCurrency)

	err := validate.Var(storeReq.Currency, "currency")
	if err != nil {
		h.Logger.Error(err)

		context.JSON(http.StatusBadRequest, gin.H{
			"error": currencyNotSupported,
		})
		return
	}

	err = validate.Struct(storeReq)
	if err != nil {
		h.Logger.Error(err)

		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	acct := accountRequestFromAccountDomain(storeReq)

	account, err := h.Service.Register(context, acct)
	if err != nil {
		h.Logger.Error(err)

		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	context.JSON(http.StatusCreated, account)
}

// clean deletes an account by ID
func (h *AccountHandler) clean(context *gin.Context) {
	id := context.Param("id")
	if id == "" {
		h.Logger.Error("no account id found")

		context.JSON(http.StatusBadRequest, gin.H{
			"error": accountIDRequired,
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
