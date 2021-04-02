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
	// "sync/atomic"
	// "unsafe"
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
	// ptr     unsafe.Pointer
)

// type GloablConfig struct {
// 	Addr                 string   `json:"addr"`
// 	Peers                []string `json:"peers"`
// 	EnableHttps          bool     `json:"enable_https"`
// 	Group                string   `json:"group"`
// 	RenameFile           bool     `json:"rename_file"`
// 	ShowDir              bool     `json:"show_dir"`
// 	Extensions           []string `json:"extensions"`
// 	RefreshInterval      int      `json:"refresh_interval"`
// 	EnableWebUpload      bool     `json:"enable_web_upload"`
// 	DownloadDomain       string   `json:"download_domain"`
// 	EnableCustomPath     bool     `json:"enable_custom_path"`
// 	Scenes               []string `json:"scenes"`
// 	AlarmReceivers       []string `json:"alarm_receivers"`
// 	DefaultScene         string   `json:"default_scene"`
// 	Mail                 Mail     `json:"mail"`
// 	AlarmUrl             string   `json:"alarm_url"`
// 	DownloadUseToken     bool     `json:"download_use_token"`
// 	DownloadTokenExpire  int      `json:"download_token_expire"`
// 	QueueSize            int      `json:"queue_size"`
// 	AutoRepair           bool     `json:"auto_repair"`
// 	Host                 string   `json:"host"`
// 	FileSumArithmetic    string   `json:"file_sum_arithmetic"`
// 	PeerId               string   `json:"peer_id"`
// 	SupportGroupManage   bool     `json:"support_group_manage"`
// 	AdminIps             []string `json:"admin_ips"`
// 	EnableMergeSmallFile bool     `json:"enable_merge_small_file"`
// 	EnableMigrate        bool     `json:"enable_migrate"`
// 	EnableDistinctFile   bool     `json:"enable_distinct_file"`
// 	ReadOnly             bool     `json:"read_only"`
// 	EnableCrossOrigin    bool     `json:"enable_cross_origin"`
// 	EnableGoogleAuth     bool     `json:"enable_google_auth"`
// 	AuthUrl              string   `json:"auth_url"`
// 	EnableDownloadAuth   bool     `json:"enable_download_auth"`
// 	DefaultDownload      bool     `json:"default_download"`
// 	EnableTus            bool     `json:"enable_tus"`
// 	SyncTimeout          int64    `json:"sync_timeout"`
// 	EnableFsnotify       bool     `json:"enable_fsnotify"`
// 	EnableDiskCache      bool     `json:"enable_disk_cache"`
// 	ConnectTimeout       bool     `json:"connect_timeout"`
// 	ReadTimeout          int      `json:"read_timeout"`
// 	WriteTimeout         int      `json:"write_timeout"`
// 	IdleTimeout          int      `json:"idle_timeout"`
// 	ReadHeaderTimeout    int      `json:"read_header_timeout"`
// 	SyncWorker           int      `json:"sync_worker"`
// 	UploadWorker         int      `json:"upload_worker"`
// 	UploadQueueSize      int      `json:"upload_queue_size"`
// 	RetryCount           int      `json:"retry_count"`
// 	SyncDelay            int64    `json:"sync_delay"`
// 	WatchChanSize        int      `json:"watch_chan_size"`
// }

// 小写的字段被认为是私有的，不会被标准的json序列化程序序列化。
type FileInfo struct {
	Filename  string `json:"fileName"`
	Fid       string `json:"fid"`
	Cid       string `json:"cid"`
	Hash      string `json:"hash"`
	Size      int64  `json:"size"`
	Path      string `json:"path"`
	Pwd       string `json:"pwd"`
}

type VerifyInfo struct {
	Fid       string `json:"fid"`
	Exit      bool `json:"exit"`
}

// func Config() *GloablConfig {
// 	return (*GloablConfig)(atomic.LoadPointer(&ptr))
// }

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

func CommonRes(c *gin.Context, code int, data map[string]interface{}, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"data": data,
		"msg": msg,
	})
}

// 多文件上传生成统一id
func GetUpId(c *gin.Context) {
	worker, err := utils.NewWorker(1)
	data := make(map[string]interface{})
	if err != nil {
		CommonRes(c, 2001, data, "生成机器id失败")
		fmt.Println("生成机器id失败",err)
		return
	}
	id := worker.GetId()
	data["fid"] = id
	CommonRes(c, 0, data, "")
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
	data := make(map[string]interface{})

	if err != nil {
		CommonRes(c, 3001, data, "chunk参数非文件类型")
		panic(err)
	}
	fmt.Println("文件信息：", saveName, hash, savePath)
	defer file.Close()

	if !CreateFileStat(saveDir) {
		CommonRes(c, 2011, data, "创建根目录失败")
		return
	}

	if !CreateFileStat(savePath) {
		CommonRes(c, 2012, data, "创建文件目录失败")
		return
	}

	out, err := os.Create(fileName)
	if err != nil {
		CommonRes(c, 2013, data, "打开目录失败")
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		CommonRes(c, 2012, data, "生成文件失败")
		return
	}
	CommonRes(c, 0, data, "")
}

func MergeFileChunk(c *gin.Context) {
	fileName := c.Query("fileName")
	hash := c.Query("hash")
	fid := c.Query("fid")	
	size := c.Query("size")
	pwd := c.Query("pwd")
	data := make(map[string]interface{})
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
			CommonRes(c, 2021, data, "打开文件目录失败")
			return
		}
		fmt.Println(ff.Name())
		b, err := ioutil.ReadAll(ff)
		if err != nil {
			fmt.Println(err)
			CommonRes(c, 2023, data, "读取文件失败")
			return
		}
		fii.Write(b)
		ff.Close()
	}
	rawCode := GetBKDRHash(hash + strconv.FormatInt(time.Now().Unix(),10))
	code := strconv.FormatUint(rawCode, 36)
	fail := InsertFid(fid)
	if fail {
		CommonRes(c, 2024, data, "插入fid数据失败")
		return
	}
	fail = InsertFileHash(hash, fileName,code, fid,size,pwd)
	if fail {
		CommonRes(c, 2025, data, "插入hash数据失败")
		return
	}
	os.RemoveAll(pathTmp)
	fii.Close()
	CommonRes(c, 0, data, "")
}

// 获取下载文件列表
func GetFileList(c *gin.Context) {
	fid := c.PostForm("fid")
	list, err := SelectFileByFid(fid)
	data := make(map[string]interface{})
	fmt.Println("获取到的文件列表", list)
	if err {
		CommonRes(c, 2031, data, "获取下载文件列表失败")
		return
	}else if len(list) == 0 {
		CommonRes(c, 2032, data, "提取码错误")
		return
	}
	data["list"] = list
	CommonRes(c, 0, data, "")
}

func Download(c *gin.Context) {
	code, _ := url.QueryUnescape(c.Query("code"))
	fmt.Print("下载文件code：", code)
	file := SelectFileByCode(code)
	
	c.Writer.WriteHeader(http.StatusOK)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(file.Filename))
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
	c.File(saveDir + "\\" + file.Hash)
}

// 验证是否存在hash
func VerifyFile(c *gin.Context) {
	hash := c.Query("hash")
	newfid := c.Query("fid")
	isExit := false
	data := make(map[string]interface{})
	fid, exist := SelectFileByHash(hash)
	fmt.Println("查询文件by hash", newfid, fid)
	if exist {
		err := UpdateFid(newfid, fid)
		if err {
			CommonRes(c, 2041, data, "更新数据失败")
			return
		}
		isExit = true
	}
	data["isExit"] = isExit
	CommonRes(c, 0, data, "")
}

// 插入上传的文件hash信息
func InsertFileHash(hash string, name string, code string, fid string,size string,pwd string) bool{
	stmt,err := DB.Prepare("insert into `file_info`(hash,filename,cid,fid,size,pwd)values(?,?,?,?,?,?)")
	if err != nil{
		fmt.Println("file_info预处理失败:",err)
		return true
	}
	result,err := stmt.Exec(hash,name,code,fid,size,pwd)
	if err != nil{
		fmt.Println("插入文件信息失败:",err)
		return true
	}else{
		rows,_ := result.RowsAffected()
		fmt.Println("执行成功,影响行数",rows,"行" )
		return false
	}
}
func InsertFid(fid string) bool{
	stmt,err := DB.Prepare("insert into `fid_collection`(fid,time)values(?,?)")
	if err != nil{
		fmt.Println("fid_collection预处理失败:",err)
		return true
	}
	t := time.Now()
	result,err := stmt.Exec(fid,t)
	if err != nil{
		fmt.Println("插入fid失败:",err)
		return true
	}else{
		rows,_ := result.RowsAffected()
		fmt.Println("执行成功,影响行数",rows,"行" )
		return false
	}
}
func UpdateFid(newfid string, oldfid string) bool{
	sqlStr := "update fid_collection set fid=?, time=? where fid=?"
	time := time.Now()
	ret, err := DB.Exec(sqlStr, newfid, time, oldfid)
	if err != nil {
		fmt.Printf("update failed, err:%v\n", err)
		return true
	}
	n, err := ret.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return true
	}
	fmt.Printf("更新成功, affected rows:%d\n", n)
	return false
}
// 查询文件hash
func SelectFileByHash(hash string) (string, bool){
	var info FileInfo
	err := DB.QueryRow("select * from file_info where hash = ?", hash).Scan(&info.Hash, &info.Filename, &info.Cid, &info.Fid, &info.Size,&info.Pwd)
	if err != nil {
		fmt.Println("查询没结果")
	}
	if info.Fid != "" {
		return info.Fid, true
	} else {
		return "", false
	}
}

// 查询提取码
func SelectFileByCode(code string) FileInfo {
	var info FileInfo
	err := DB.QueryRow("select * from file_info where cid = ?", code).Scan(&info.Hash, &info.Filename, &info.Cid, &info.Fid, &info.Size, &info.Pwd)
	if err != nil {
		fmt.Println("查询没结果")
	}
	return info
}

// 查询文件列表
func SelectFileByFid(fid string) ([]map[string]string, bool) {
	rows,err := DB.Query("select * from file_info where fid = ?", fid)
	ret := make([]map[string]string, 0)
	defer rows.Close() 
	if err != nil {
		fmt.Println("查询出错")
		return ret, true
	}
	columns, _ := rows.Columns()            //获取列的信息
	count := len(columns)
	var values = make([]interface{}, count) //创建一个与列的数量相当的空接口
	for i, _ := range values {
		var ii interface{} //为空接口分配内存
		values[i] = &ii    //取得这些内存的指针，因后继的Scan函数只接受指针
	}
	for rows.Next() {
		err := rows.Scan(values...)
		m := make(map[string]string) //用于存放1列的 [键/值] 对
		if err != nil {
			panic(err)
		}
		for i, colName := range columns {
			var raw_value = *(values[i].(*interface{})) //读出raw数据，类型为byte
			b, _ := raw_value.([]byte)
			v := string(b) //将raw数据转换成字符串
			m[colName] = v //colName是键，v是值
		}
		ret = append(ret, m) //将单行所有列的键值对附加在总的返回值上（以行为单位）
	}
	return ret, false
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