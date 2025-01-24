package utils

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

func ParsePaginationParams(c *gin.Context) (int, int) {
	pnStr := c.DefaultQuery("pageNo", "1")
	pSizeStr := c.DefaultQuery("pageSize", "9999")

	pn, err := strconv.Atoi(pnStr)
	if err != nil || pn < 1 {
		pn = 1
	}

	pSize, err := strconv.Atoi(pSizeStr)
	if err != nil || pSize < 1 {
		pSize = 10 // 如果转换失败或者每页大小小于1，设置默认值为10
	}

	return pn, pSize
}

func DefaultQueryInt(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.DefaultQuery(key, "")

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func DefaultQueryBool(c *gin.Context, key string, defaultValue bool) bool {
	valueStr := c.DefaultQuery(key, "true")

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
