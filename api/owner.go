package api

import (
	"github.com/Drelf2018/webhook"
	"github.com/Drelf2018/webhook/utils"
	"github.com/gin-gonic/gin"
)

func GetExecute(ctx *gin.Context) (any, error) {
	_, keep := ctx.GetQuery("keep")
	return utils.Shell(ctx.Query("cmd"), ctx.Query("dir"), keep)
}

func GetShutdown(ctx *gin.Context) (any, error) {
	err := CloseDB()
	if err != nil {
		return 1, err
	}
	webhook.Shutdown()
	return "人生有梦，各自精彩！", nil
}

	}
	}
}
