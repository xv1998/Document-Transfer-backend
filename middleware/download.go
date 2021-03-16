package middleware

// import (
// 	"fmt"

// 	"github.com/gin-gonic/gin"
// )

// func Download(c *gin.Context) {
// 	fileName := c.Query("fileName")
// 	fmt.Print("下载文件名称：", fileName)
// 	c.Header("Content-Type", "application/octet-stream")
// 	c.Header("Content-Disposition", "attachment; filename="+fileName)
// 	c.Header("Content-Transfer-Encoding", "binary")
// 	c.Header("Cache-Control", "no-cache")
// 	c.File(filePath)
// }
