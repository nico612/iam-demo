package user

import (
	"github.com/gin-gonic/gin"
	"github.com/marmotedu/component-base/pkg/auth"
	"github.com/marmotedu/component-base/pkg/core"
	metav1 "github.com/marmotedu/component-base/pkg/meta/v1"
	"github.com/marmotedu/errors"
	"github.com/nico612/iam-demo/internal/pkg/code"
	"github.com/nico612/iam-demo/pkg/log"
)

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"omitempty"`
	NewPassword string `json:"newPassword" binding:"password"`
}

func (u *UserController) ChangePassword(c *gin.Context) {
	log.L(c).Infof("change password function called.")

	var r ChangePasswordRequest

	if err := c.ShouldBindJSON(&r); err != nil {
		core.WriteResponse(c, errors.WithCode(code.ErrBind, err.Error()), nil)

		return
	}

	user, err := u.srv.Users().Get(c, c.Param("name"), metav1.GetOptions{})
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	if err := user.Compare(r.OldPassword); err != nil {
		core.WriteResponse(c, errors.WithCode(code.ErrPasswordIncorrect, err.Error()), nil)

		return
	}

	user.Password, _ = auth.Encrypt(r.NewPassword)
	if err := u.srv.Users().ChangePassword(c, user); err != nil {
		core.WriteResponse(c, err, nil)
	}

	core.WriteResponse(c, nil, nil)

}
