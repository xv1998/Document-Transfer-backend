/**
* 处理文件
**/
package middleware

import (
	"backend/utils"
	"database/sql"
	"fmt"
	"io"
	"time"
	"strconv"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-gonic/gin"
)

//数据库配置
const (
	userName = "root"
	password = "123456"
	ip       = "127.0.0.1"
	port     = "3306"
	dbName   = "tomato"
)

//Db数据库连接池
var DB *sql.DB

var (
	saveDir = utils.GetParentPath() + "\\attachment" // 文件存储根目录
)

type FileInfo struct {
	filename string
	hash string
	code string
}

//注意方法名大写，就是public
func InitDB() {
	//构建连接："用户名:密码@tcp(IP:端口)/数据库?charset=utf8"
	path := strings.Join([]string{userName, ":", password, "@tcp(", ip, ":", port, ")/", dbName, "?charset=utf8"}, "")
	//打开数据库,前者是驱动名，所以要导入： _ "github.com/go-sql-driver/mysql"
	DB, _ = sql.Open("mysql", path)
	//设置数据库最大连接数
	DB.SetConnMaxLifetime(100)
	//设置上数据库最大闲置连接数
	DB.SetMaxIdleConns(10)
	//验证连接
	if err := DB.Ping(); err != nil {
		fmt.Println("open database fail")
		return
	}
	fmt.Println("connnect success")
}

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
	chunkhash := c.PostForm("chunkhash")
	savePath = saveDir + "\\" + hash + "-collection"
	fileName := savePath + "\\" + chunkhash

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
	fileName := c.Query("fileName")
	hash := c.Query("hash")
	pathTmp := saveDir + "\\" + hash + "-collection"
	path := saveDir + "\\" + hash
	files, _ := ioutil.ReadDir(pathTmp)
	fii, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Sprintln("open file path:", path, "error:", err))
	}
	defer fii.Close()
	for _, f := range files {
		ff, err := os.OpenFile(pathTmp+"\\"+f.Name(), os.O_RDONLY, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(ff.Name())
		b, err := ioutil.ReadAll(ff)
		if err != nil {
			fmt.Println(err)
			return
		}
		fii.Write(b)
		ff.Close()
	}
	rawCode := GetBKDRHash(hash + strconv.FormatInt(time.Now().Unix(),10))
	code := strconv.FormatUint(rawCode, 36)
	InsertFileHash(hash, fileName,code)
	os.RemoveAll(pathTmp)
	fii.Close()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"code": code,
		},
	})
}

func GetFileList(c *gin.Context) {
	code := c.PostForm("code")
	file := SelectFileByCode(code)
	fmt.Println("查询code对应的文件", file, file.code)
	// arr := [5]float32{1000.0, 2.0, 3.4, 7.0, 50.0}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list": file,
		},
	})
}

func Download(c *gin.Context) {
	fileName, _ := url.QueryUnescape(c.Query("fileName"))
	fmt.Print("下载文件名称：", fileName)
	// file, err := os.Open(saveDir + "\\" + fileName)
	// if err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{
	// 		"success": false,
	// 		"message": "文件加载失败:",
	// 	})
	// 	return
	// }
	// defer file.Close()
	c.Writer.WriteHeader(http.StatusOK)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")
	// c.File(file)
	// _, err = io.Copy(c.Writer, file)
	// if err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{
	// 		"success": false,
	// 		"message": "文件加载失败:",
	// 	})
	// 	return
	// }
	// c.Writer.Write([]byte(file))
	c.File(saveDir + "\\" + fileName)
}

// 验证是否存在hash
func VerifyFile(c *gin.Context) {
	hash := c.Query("hash")
	fmt.Println("验证hash", hash)
	isExit := SelectFileByHash(hash)
	c.JSON(http.StatusOK, gin.H{
		"code":   0,
		"isExit": isExit,
	})
}

// 插入上传的文件hash信息
func InsertFileHash(hash string, name string, code string) {
	stmt,err := DB.Prepare("insert into `file_info`(hash,filename,code)values(?,?,?)")
	if err != nil{
		fmt.Println("预处理失败:",err)
	}
	result,err := stmt.Exec(hash,name,code)
	if err != nil{
		fmt.Println("执行预处理失败:",err)
	}else{
		rows,_ := result.RowsAffected()
		fmt.Println("执行成功,影响行数",rows,"行" )
	}
}

// 查询文件hash
func SelectFileByHash(hash string) bool {
	var info FileInfo
	err := DB.QueryRow("select * from file_info where hash = ?", hash).Scan(&info.hash, &info.filename, &info.code)
	if err != nil {
		fmt.Println("查询没结果")
	}
	if info.hash != "" {
		return true
	} else {
		return false
	}
}

// 查询提取码
func SelectFileByCode(code string) FileInfo {
	var info FileInfo
	err := DB.QueryRow("select * from file_info where code = ?", code).Scan(&info.hash, &info.filename, &info.code)
	if err != nil {
		fmt.Println("查询没结果")
	}
	return info
}

// BKDR Hash
func GetBKDRHash(s string) uint64{
	seed := uint64(131)
	hash := uint64(0)
	for i := 0; i < len(s); i++ {
		hash = (hash * seed) + uint64(s[i])
	}
	return hash & 0x7FFFFFFF
}