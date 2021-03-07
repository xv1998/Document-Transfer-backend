/**
* 处理文件
**/
package middleware

import (
	"backend/utils"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var (
	saveDir = utils.GetParentPath() + "\\attachment" // 文件存储根目录
)

// const (
// 	STORE_DIR_NAME = ""
// )

// 初始化存储目录
func CreateFileStat(saveDir string) bool {
	//打开目录
	localFileInfo, fileStatErr := os.Stat(saveDir)

	//目录不存在
	if fileStatErr != nil || !localFileInfo.IsDir() {
		//创建目录
		errByMkdirAllDir := os.MkdirAll(saveDir, 0755)
		if errByMkdirAllDir != nil {
			return false
		}
	}

	return true
}

func FileUpload(c *gin.Context) {
	// fileNameParam := c.PostForm("filename") // 文件名称
	// fmt.Println("文件名", fileNameParam)
	//存储目录
	// var saveDir = utils.GetParentPath() + "\\attachment"
	// 文件目录名称
	var saveName = ""
	// 保存的文件夹名称
	var savePath = ""
	// FormFile方法会读取参数“upload”后面的文件名，返回值是一个File指针，和一个FileHeader指针，和一个err错误。
	file, _, err := c.Request.FormFile("chunk")
	saveName = c.PostForm("filename")
	hash := c.PostForm("hash")
	savePath = saveDir + "\\" + saveName + "collection"
	fileName := savePath + "\\" + hash

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 101,
		})
		panic(err)
	}
	fmt.Println("文件信息：", saveName, hash, savePath)
	defer file.Close()
	// //获取上传文件的后缀(类型)
	// uploadFileNameWithSuffix := path.Base(header.Filename)
	// uploadFileType := path.Ext(uploadFileNameWithSuffix)

	// 保存的文件夹名称
	// saveName = fileNameParam + uploadFileType
	// savePath = saveDir + "\\" + header.Filename

	if !CreateFileStat(saveDir) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"path":    saveDir,
			"message": "创建根目录失败",
		})
	}

	if !CreateFileStat(savePath) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"path":    savePath,
			"message": "创建文件目录失败",
		})
	}

	out, err := os.Create(fileName)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "打开目录失败",
		})
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "生成文件失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
	})
}

func MergeFileChunk(c *gin.Context) {
	saveName := c.Query("fileName")
	pathTmp := saveDir + "\\" + saveName + "collection"
	path := saveDir + "\\" + saveName
	files, _ := ioutil.ReadDir(pathTmp)
	fii, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintln("open file path:", path, "error:", err))
	}
	defer fii.Close()
	for _, f := range files {
		ff, err := os.OpenFile(f.Name(), os.O_RDONLY, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return
		}
		b, err := ioutil.ReadAll(ff)
		if err != nil {
			fmt.Println(err)
			return
		}
		fii.Write(b)
		ff.Close()
	}
	os.RemoveAll(pathTmp)
	fii.Close()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
	})
}
