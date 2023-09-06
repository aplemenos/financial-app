package healthchecks

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HealthcheckHandler struct {
	Service Service

	Logger *zap.SugaredLogger
}

// router sets up all the routes for healthcheck service
func (h *HealthcheckHandler) Router(routerGroup *gin.RouterGroup) {
	routerGroup.GET("/alive", h.aliveCheck)
}

func (h *HealthcheckHandler) aliveCheck(context *gin.Context) {
	if err := h.Service.Alive(context); err != nil {
		h.Logger.Error(err)
		context.JSON(http.StatusServiceUnavailable, gin.H{
			"error": err.Error(),
		})

		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message": "I am Alive!",
	})
}
