package user

import (
	"github.com/gin-gonic/gin"
	"github.com/marmotedu/component-base/pkg/core"
	metav1 "github.com/marmotedu/component-base/pkg/meta/v1"
	"github.com/nico612/iam-demo/pkg/log"
)

func (u *UserController) DeleteCollection(c *gin.Context) {
	log.L(c).Infof("batch delete user function called.")
	usernames := c.QueryArray("name")

	if err := u.srv.Users().DeleteCollection(c, usernames, metav1.DeleteOptions{}); err != nil {
		core.WriteResponse(c, err, nil)

		return
	}

	core.WriteResponse(c, nil, nil)
}
