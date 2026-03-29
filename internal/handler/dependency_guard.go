package handler

import (
	"net/http"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

// ensureServiceReady 统一处理 Handler 依赖为空时的响应。
func ensureServiceReady(c *gin.Context, dependency any) bool {
	if dependency != nil {
		return true
	}

	c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: service.ErrInternal.Error()})
	return false
}
