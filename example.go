package main

import (
	"backend/middleware"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
)

func InitPeriodicTask(f func()){
	c := cron.New()
	spec := "0 0 */1 * * ?"
	c.AddFunc(spec, f)
	c.Start()
}

func main() {
	go middleware.InitDB()
	// go InitPeriodicTask(middleware.ClearDisk)
	r := gin.Default()
	r.Use(middleware.Cors())
	r.POST("/file/upload", middleware.FileUpload)
	r.GET("/file/getupid", middleware.GetUpId)
	r.GET("/file/merge", middleware.MergeFileChunk)
	r.GET("/file/download", middleware.Download)
	r.GET("/file/verify", middleware.VerifyFile)
	r.POST("/file/checkpwd", middleware.CheckPwd)
	r.Run(":8081")
}

// package main
 
// import (
//     "github.com/robfig/cron"
//     "log"
// )
 
// func main() {
//     i := 0
//     c := cron.New()
//     spec := "*/5 * * * * ?"
//     c.AddFunc(spec, func() {
//         i++
//         log.Println("cron running:", i)
//     })
//     c.Start()
 
//     select{}
// }