package authorize

import (
	"github.com/gin-gonic/gin"
	"github.com/marmotedu/component-base/pkg/core"
	"github.com/marmotedu/errors"
	"github.com/nico612/iam-demo/internal/authzserver/authorization"
	"github.com/nico612/iam-demo/internal/authzserver/authorization/authorizer"
	"github.com/nico612/iam-demo/internal/pkg/code"
	"github.com/ory/ladon"
)

// AuthzController create a authorizer handler used to handle authorizer request.
type AuthzController struct {
	store authorizer.PolicyGetter
}

// NewAuthzController creates a authorizer handler.
func NewAuthzController(store authorizer.PolicyGetter) *AuthzController {
	return &AuthzController{
		store: store,
	}
}

// Authorize returns whether a request is allow or deny to access a resource and do some action
// under specified condition.
func (a *AuthzController) Authorize(c *gin.Context) {
	var r ladon.Request
	if err := c.ShouldBind(&r); err != nil {
		core.WriteResponse(c, errors.WithCode(code.ErrBind, err.Error()), nil)

		return
	}

	auth := authorization.NewAuthorizer(authorizer.NewAuthorization(a.store))
	if r.Context == nil {
		r.Context = ladon.Context{}
	}

	r.Context["username"] = c.GetString("username")
	rsp := auth.Authorize(&r)

	core.WriteResponse(c, nil, rsp)
}
