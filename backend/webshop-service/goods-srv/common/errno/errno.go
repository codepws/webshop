// errno/errno.go

package errno

import (
	"encoding/json"
)

var _ Error = (*err)(nil)

type Error interface {
	// i 为了避免被其他包实现
	i()
	// WithData 设置成功时返回的数据
	WithData(data interface{}) Error
	// WithID 设置当前请求的唯一ID
	WithID(id string) Error
	// ToString 返回 JSON 格式的错误详情
	ToString() string

	// WithItems
	WithItems(items interface{}) Error
	// WithCount
	WithCount(count int64) Error

	WithSysError(err error) Error
}

type err struct {
	Code int         `json:"code"`           // 业务编码
	Msg  string      `json:"msg"`            // 错误描述
	Data interface{} `json:"data,omitempty"` // 成功时返回的数据
	ID   string      `json:"id,omitempty"`   // 当前请求的唯一ID，便于问题定位，忽略也可以

	Items interface{} `json:"items"`
	Count int64       `json:"count"`

	SysErr error `json:"-"`
}

func NewError(code int, msg string) Error {
	return &err{
		Code:   code,
		Msg:    msg,
		Data:   nil,
		Items:  nil,
		Count:  0,
		SysErr: nil,
	}
}

func (e *err) i() {}

func (e *err) WithSysError(err error) Error {
	e.SysErr = err
	return e
}

func (e *err) WithData(data interface{}) Error {
	e.Data = data
	return e
}

func (e *err) WithID(id string) Error {
	e.ID = id
	return e
}

func (e *err) WithItems(items interface{}) Error {
	e.Items = items
	return e
}

func (e *err) WithCount(count int64) Error {
	e.Count = count
	return e
}

// ToString 返回 JSON 格式的错误详情
func (e *err) ToString() string {
	var syserr string
	if e.SysErr != nil {
		syserr = e.SysErr.Error()
	}
	err := &struct {
		Code   int         `json:"code"`
		Msg    string      `json:"msg"`
		Data   interface{} `json:"data"`
		ID     string      `json:"id,omitempty"`
		SysErr string      //`json:"-"`
	}{
		Code:   e.Code,
		Msg:    e.Msg,
		Data:   e.Data,
		ID:     e.ID,
		SysErr: syserr,
	}

	raw, _ := json.Marshal(err)
	return string(raw)
}
