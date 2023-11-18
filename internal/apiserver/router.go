package apiserver

import (
	"github.com/gin-gonic/gin"
	"github.com/marmotedu/component-base/pkg/core"
	"github.com/marmotedu/errors"
	"github.com/nico612/iam-demo/internal/apiserver/controller/v1/user"
	"github.com/nico612/iam-demo/internal/apiserver/store/mysql"
	"github.com/nico612/iam-demo/internal/pkg/code"
	"github.com/nico612/iam-demo/internal/pkg/middleware"
	"github.com/nico612/iam-demo/internal/pkg/middleware/auth"
)

func initRouter(g *gin.Engine) {
	installMiddleware(g)
	installController(g)
}

func installMiddleware(g *gin.Engine) {

}

func installController(g *gin.Engine) *gin.Engine {

	// Middlewares.
	jwtStrategy, _ := newJWTAuth().(auth.JWTStrategy)
	g.POST("/login", jwtStrategy.LoginHandler)
	g.POST("/logout", jwtStrategy.LogoutHandler)
	g.POST("refresh", jwtStrategy.RefreshHandler)

	auto := newJWTAuth()

	// 404
	g.NoRoute(auto.AuthFunc(), func(c *gin.Context) {
		core.WriteResponse(c, errors.WithCode(code.ErrPageNotFound, "Page not found."), nil)
	})

	storeIns, _ := mysql.GetMySQLFactoryOr(nil)
	v1 := g.Group("/v1")
	{
		// user RESTful resource
		userv1 := v1.Group("/users")
		{
			userController := user.NewUserController(storeIns)
			userv1.POST("", userController.Create)
			userv1.Use(auto.AuthFunc(), middleware.Validation())

			userv1.DELETE("", userController.DeleteCollection)
			userv1.PUT(":name/change-password", userController.ChangePassword)
			userv1.PUT(":name", userController.Update)
			userv1.GET("", userController.List)
			userv1.GET(":name", userController.Get) // admin api
		}

		v1.Use(auto.AuthFunc())

	}

	return g
}
