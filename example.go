package main

import (
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(middleware.Cors())
	r.POST("/file/upload", middleware.FileUpload)
	r.GET("/file/merge", middleware.MergeFileChunk)
	r.Run(":8081")
}
