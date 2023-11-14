package code

import (
	"github.com/marmotedu/errors"
	"github.com/novalagung/gubrak"
)

type ErrCode struct {
	// C 对应的 code 码
	C int

	// HTTP 关联的 code http 状态
	HTTP int

	// 面向外部（用户）的错误文本。
	Ext string

	// Ref 指定的参考文档
	Ref string
}

func (coder ErrCode) Code() int {
	return coder.C
}

func (coder ErrCode) String() string {
	return coder.Ext
}

func (coder ErrCode) Reference() string {
	return coder.Ref
}

// HTTPStatus 返回关联的 HTTP 状态代码（如果有）。 否则，返回 200
func (coder ErrCode) HTTPStatus() int {
	if coder.HTTP == 0 {
		return 500
	}
	return coder.HTTP
}

func register(code int, httpStatus int, message string, refs ...string) {
	found, _ := gubrak.Includes([]int{200, 400, 401, 403, 404, 500}, httpStatus)
	if !found {
		panic("http code not in `200, 400, 401, 403, 404, 500`")
	}

	var reference string
	if len(refs) > 0 {
		reference = refs[0]
	}

	coder := &ErrCode{
		C:    code,
		HTTP: httpStatus,
		Ext:  message,
		Ref:  reference,
	}

	errors.MustRegister(coder)
}
