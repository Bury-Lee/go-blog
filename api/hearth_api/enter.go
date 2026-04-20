package hearth_api

import (
	"StarDreamerCyberNook/common/response"

	"github.com/gin-gonic/gin"
)

type HearthApi struct {
}

func (h *HearthApi) Hearth(ctx *gin.Context) {
	response.OkWithMsg("存活", ctx)
}
