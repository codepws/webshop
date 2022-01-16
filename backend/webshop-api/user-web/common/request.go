package common

import (
	"github.com/gin-gonic/gin"
)

func getCurrentUserID(c *gin.Context) (userID uint64, err error) {

	userID = 0
	err = nil
	/*
		_userID, ok := c.Get(controller.KeyCtxContextUserID)
		if !ok {
			//err = ErrorUserNotLogin
			return
		}
		userID, ok = _userID.(uint64)
		if !ok {
			//err = ErrorUserNotLogin
			return
		}
	*/
	return
}
