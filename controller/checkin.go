package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

// DoCheckin handles the check-in request
// POST /api/user/checkin
// Requirements: 2.1, 2.2, 2.3
func DoCheckin(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请先登录",
		})
		return
	}

	result, err := service.DoCheckin(userId)
	if err != nil {
		common.SysLog("checkin error for user " + string(rune(userId)) + ": " + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "签到失败，请稍后重试",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": result.Success,
		"message": result.Message,
		"data":    result,
	})
}

// GetCheckinStatus handles the check-in status request
// GET /api/user/checkin/status
// Requirements: 2.5
func GetCheckinStatus(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "请先登录",
		})
		return
	}

	status, err := service.GetCheckinStatus(userId)
	if err != nil {
		common.SysLog("get checkin status error for user " + string(rune(userId)) + ": " + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取签到状态失败，请稍后重试",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    status,
	})
}
