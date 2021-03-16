package main

import (
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	middleware.InitDB()
	r := gin.Default()
	r.Use(middleware.Cors())
	r.POST("/file/upload", middleware.FileUpload)
	r.GET("/file/merge", middleware.MergeFileChunk)
	r.GET("/file/download", middleware.Download)
	r.GET("/file/verify", middleware.VerifyFile)
	r.POST("/file/getlist", middleware.GetFileList)
	r.Run(":8081")
}
