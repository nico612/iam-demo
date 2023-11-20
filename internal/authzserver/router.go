package authzserver

import (
	"github.com/gin-gonic/gin"
	"github.com/marmotedu/component-base/pkg/core"
	"github.com/marmotedu/errors"
	"github.com/nico612/iam-demo/internal/authzserver/controller/v1/authorize"
	"github.com/nico612/iam-demo/internal/authzserver/load/cache"
	"github.com/nico612/iam-demo/internal/pkg/code"
	"github.com/nico612/iam-demo/pkg/log"
)

func installMiddleware(g *gin.Engine) {
}

func installController(g *gin.Engine) *gin.Engine {
	auth := newCacheAuth() // 缓存中间件

	g.NoRoute(auth.AuthFunc(), func(c *gin.Context) {
		core.WriteResponse(c, errors.WithCode(code.ErrPageNotFound, "page not found."), nil)
	})

	cacheIns, _ := cache.GetCacheInsOr(nil)
	if cacheIns == nil {
		log.Panicf("get nil cache instance")
	}

	apiv1 := g.Group("/v1", auth.AuthFunc())
	{
		authzController := authorize.NewAuthzController(cacheIns)

		// Router for authorization
		apiv1.POST("/authz", authzController.Authorize)
	}

	return g
}
