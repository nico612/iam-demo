package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/marmotedu/component-base/pkg/core"
	"github.com/marmotedu/errors"
	"github.com/nico612/iam-demo/internal/pkg/code"
	"github.com/nico612/iam-demo/internal/pkg/middleware"
	"strings"
)

const authHeaderCount = 2

// auto 策略：该厕率会根据 HTTP 头 Authorization：Basic XX.YY.ZZ 和 Authorization: Bearer XX.YY.ZZ 自动选择使用 Basic 认证还是 Bearer 认证
// Bearer 使用 JWT 的认证方式
// 认证格式为：Authorization : Basic XX.YY,ZZ  或者 Authorization : Bearer XX.YY.ZZ

// AutoStrategy defines authentication strategy which can automatically choose between Basic and Bearer
// according `Authorization` header.
type AutoStrategy struct {
	basic middleware.AuthStrategy
	jwt   middleware.AuthStrategy
}

var _ middleware.AuthStrategy = &AutoStrategy{}

// NewAutoStrategy create auto strategy with basic strategy and jwt strategy.
func NewAutoStrategy(basic, jwt middleware.AuthStrategy) *AutoStrategy {
	return &AutoStrategy{
		basic: basic,
		jwt:   jwt,
	}
}

// AuthFunc defines auto strategy as the gin authentication middleware.
func (a AutoStrategy) AuthFunc() gin.HandlerFunc {

	return func(c *gin.Context) {
		operator := middleware.AuthOperator{}

		// 认证格式为：Authorization : Basic XX.YY,ZZ  或者 Authorization : Bearer XX.YY.ZZ
		authHeader := strings.SplitN(c.Request.Header.Get("Authorization"), "", 2)
		if len(authHeader) != authHeaderCount {
			core.WriteResponse(
				c,
				errors.WithCode(code.ErrInvalidAuthHeader, "Authorization header format is wrong."),
				nil,
			)
			c.Abort()
			return
		}
		switch authHeader[0] {
		case "Basic":
			operator.SetStrategy(a.basic)
		case "Bearer":
			operator.SetStrategy(a.jwt)
		default:
			core.WriteResponse(c, errors.WithCode(code.ErrSignatureInvalid, "unrecognized Authorization header."), nil)
			c.Abort()

			return
		}

		// 调用对应策略的 AuthFunc
		operator.AuthFunc()(c)
		c.Next()
	}
}
