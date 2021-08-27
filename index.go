package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var programConfig map[string]interface{}
var localConfig map[string]interface{}

func commandLinePreProcess(args []string) {

	for i := 0; i < len(args); i++ {
		if args[i][0] == '-' {
			if i+1 == len(args) {
				programConfig[args[i][1:]] = true
			} else {
				if args[i+1][0] == '-' {
					programConfig[args[i][1:]] = true
				} else {
					programConfig[args[i][1:]] = args[i+1]
					i += 1
				}
			}
		}
	}
}

func readFile(filePath string) string {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
	}

	// 要记得关闭
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	return string(byteValue)
}

func checkExsit(filepath string) bool {
	_, err := os.Stat(filepath)
	if err == nil {
		return true
	} else {
		return false
	}
}

func readJson(filepath string) map[string]interface{} {
	filecontent := []byte(readFile("config.json"))
	var result map[string]interface{}
	json.Unmarshal(filecontent, &result)
	return result
}

func configCheck(filepath string) {
	if !checkExsit(filepath) {
		f, err := os.Create(filepath)
		defer f.Close()
		if err != nil {
			// 创建文件失败处理
			log.Fatal("Config文件创建失败，请检查目录权限")
		} else {
			_, err = f.Write([]byte("{}"))
			if err != nil {
				log.Fatal("Config文件创建失败，请检查目录权限")
			}
		}
	} else {
		localConfig = readJson("config.json")
	}
}

func getFileInfo(filepath string) {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		log.Fatal("文件错误")
	}
	fmt.Println(fileInfo)
}

func refresh_request(refresh_token string) map[string]interface{} {
	url := "https://api.aliyundrive.com/token/refresh"
	method := "POST"
	requ := fmt.Sprintf(`{"refresh_token":"%s"}`, refresh_token)
	payload := strings.NewReader(requ)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return make(map[string]interface{})
	}
	req.Header.Add("Content-Type", "application/json;charset=utf-8")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return make(map[string]interface{})
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return make(map[string]interface{})
	}
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return result
}

func getUploadData(refresh_token string) map[string]interface{} {
	result := refresh_request(refresh_token)
	if result["status"] == nil {
		log.Fatal("Refresh Token 已过期，请重新获取")
	}
	if result["status"].(string) == "enabled" {
		return result
	} else {
		log.Fatal("Refresh Token无效,请重新获取")
		return make(map[string]interface{})
	}
}

func writeFile(filepath string, content string) {
	file, err := os.OpenFile(filepath, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}
	//及时关闭file句柄
	defer file.Close()
	//写入文件时，使用带缓存的 *Writer
	write := bufio.NewWriter(file)
	write.WriteString(content)
	//Flush将缓存的文件真正写入到文件中
	write.Flush()
}

func updateRefreshToken(refresh_token string) {
	localConfig["refresh_token"] = refresh_token
	result, _ := json.Marshal(localConfig)
	writeFile("config.json", string(result))
}

func getUploadInfo(fileName string, fileSize string, fileSha1 string) []string {
	filePath := programConfig["filePath"]
	var authKey string
	var ParentId string
	if (filePath == nil || filePath == true) && !(programConfig["action"] == "server") {
		log.Fatal("未选择文件,-filePath 文件绝对路径")
	}
	if programConfig["refreshToken"] == nil {
		if localConfig["refresh_token"] == nil {
			log.Fatal("Refresh Token未配置")
		} else {
			authKey = localConfig["refresh_token"].(string)
		}
	} else {
		authKey = programConfig["refreshToken"].(string)
	}
	if programConfig["ParentId"] == nil {
		if localConfig["ParentId"] == nil {
			log.Fatal("错误,未配置上传目录")
		} else {
			ParentId = localConfig["ParentId"].(string)
			fmt.Println("ParentId:" + ParentId)
		}
	} else {
		ParentId = programConfig["ParentId"].(string)
		fmt.Println("ParentId:" + ParentId)
	}
	fmt.Println("尝试获取Token")
	part_number, _ := strconv.Atoi(fileSize)
	part_number = part_number / 10485760
	uploadData := getUploadData(authKey)
	updateRefreshToken(uploadData["refresh_token"].(string))
	drive_id := uploadData["default_drive_id"].(string)
	var urls []string
	fmt.Println("准备上传")
	for i := 1; i <= 1; i++ {
		url := "https://api.aliyundrive.com/adrive/v2/file/createWithFolders"
		method := "POST"

		fileConfig := fmt.Sprintf(`{"drive_id":"%s","part_info_list":[{"part_number":%d}],"parent_file_id":"%s","name":"%s","type":"file","check_name_mode":"auto_rename","size":%s}`, drive_id, i, ParentId, fileName, fileSize)
		payload := strings.NewReader(fileConfig)

		client := &http.Client{}
		req, err := http.NewRequest(method, url, payload)

		if err != nil {
			fmt.Println(err)

		}
		req.Header.Add("Authorization", uploadData["access_token"].(string))
		req.Header.Add("Content-Type", "text/plain")
		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
		}
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		uploadurl := result["part_info_list"].([]interface{})[0].(map[string]interface{})["upload_url"].(string)
		urls = append(urls, uploadurl)
		urls = append(urls, result["upload_id"].(string))
		urls = append(urls, result["file_id"].(string))
	}

	return urls
}

func Process() {
	action := programConfig["action"]
	if action == nil {
		log.Fatal("请选择操作,-action 操作")
	}
	switch action {
	case "GetUploadInfo":
		break
	case "localUpload":
		localUpload()
		break
	case "download":
		downloadProcess()
		break
	case "server":
		server()
		break
	}
}

func completeRequest(upload_id string, file_id string) string {
	filePath := programConfig["filePath"]
	var authKey string
	var ParentId string
	if (filePath == nil || filePath == true) && !(programConfig["action"] == "server") {
		log.Fatal("未选择文件,-filePath 文件绝对路径")
	}
	if programConfig["refreshToken"] == nil {
		if localConfig["refresh_token"] == nil {
			log.Fatal("Refresh Token未配置")
		} else {
			authKey = localConfig["refresh_token"].(string)
		}
	} else {
		authKey = programConfig["refreshToken"].(string)
	}
	if programConfig["ParentId"] == nil {
		if localConfig["ParentId"] == nil {
			log.Fatal("错误,未配置上传目录")
		} else {
			ParentId = localConfig["ParentId"].(string)
			fmt.Println("ParentId:" + ParentId)
		}
	} else {
		ParentId = programConfig["ParentId"].(string)
		fmt.Println("ParentId:" + ParentId)
	}
	url := "https://api.aliyundrive.com/v2/file/complete"
	method := "POST"
	uploadData := getUploadData(authKey)
	updateRefreshToken(uploadData["refresh_token"].(string))
	drive_id := uploadData["default_drive_id"].(string)
	payload := strings.NewReader(fmt.Sprintf(`{"drive_id":"%s","upload_id":"%s","file_id":"%s"}`, drive_id, upload_id, file_id))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("Authorization", uploadData["access_token"].(string))
	req.Header.Add("Content-Type", "text/plain")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(body)
}

func uploadFunc(c *gin.Context) {
	fileName := c.PostForm("fileName")
	fileSize := c.PostForm("fileSize")
	fileSha1 := c.PostForm("fileSha")
	result := getUploadInfo(fileName, fileSize, fileSha1)
	c.JSON(200, gin.H{
		"result": result,
	})
}

func completeFunc(c *gin.Context) {
	upload_id := c.PostForm("uploadId")
	file_id := c.PostForm("fileId")
	result := completeRequest(upload_id, file_id)
	var jsonResult map[string]interface{}
	json.Unmarshal([]byte(result), &jsonResult)
	c.JSON(200, gin.H{
		"result": jsonResult,
	})
}

func uploadToOSS(filepath string, uploadInfo []string) {
	content := readFile(filepath)
	url := uploadInfo[0]
	uploadId := uploadInfo[1]
	fileId := uploadInfo[2]
	method := "PUT"

	payload := strings.NewReader(content)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	completeRequest(uploadId, fileId)
}

func directUploadFunc(c *gin.Context) {
	_, header, _ := c.Request.FormFile("file")
	randnum := ""
	for i := 0; i < 10; i++ {
		randnum += fmt.Sprintf("%d", rand.Intn(10))
	}
	c.SaveUploadedFile(header, "temp"+randnum)
	fileSize := fmt.Sprintf("%d", header.Size)
	info := getUploadInfo(header.Filename, fileSize, "")
	uploadToOSS("temp"+randnum, info)
	os.Remove("temp" + randnum)
	c.JSON(200, gin.H{
		"data":   "success",
		"fileid": info[2],
	})
}

func localUpload() {
	filepath := programConfig["filePath"].(string)
	fileinfo, _ := os.Stat(filepath)
	fileName := fileinfo.Name()
	fileSize := fmt.Sprintf("%d", fileinfo.Size())
	info := getUploadInfo(fileName, fileSize, "")
	uploadToOSS(filepath, info)
	fmt.Println("Upload successfully .")
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") //请求头部
		if origin != "" {
			//接收客户端发送的origin （重要！）
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			//服务器支持的所有跨域请求的方法
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
			//允许跨域设置可以返回其他子段，可以自定义字段
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			// 允许浏览器（客户端）可以解析的头部 （重要）
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			//设置缓存时间
			c.Header("Access-Control-Max-Age", "172800")
			//允许客户端传递校验信息比如 cookie (重要)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		//允许类型校验
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok!")
		}

		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic info is: %v", err)
			}
		}()
		c.Next()
	}
}

// 文件列表
func listFiles(fileid string) map[string]interface{} {
	var authKey string
	if programConfig["refreshToken"] == nil {
		if localConfig["refresh_token"] == nil {
			log.Fatal("Refresh Token未配置")
		} else {
			authKey = localConfig["refresh_token"].(string)
		}
	} else {
		authKey = programConfig["refreshToken"].(string)
	}
	uploadData := getUploadData(authKey)
	updateRefreshToken(uploadData["refresh_token"].(string))
	drive_id := uploadData["default_drive_id"].(string)

	url := "https://api.aliyundrive.com/adrive/v3/file/list"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{"drive_id":"%s","parent_file_id":"%s","limit":200,"all":false,"url_expire_sec":1600,"image_thumbnail_process":"image/resize,w_400/format,jpeg","image_url_process":"image/resize,w_1920/format,jpeg","video_thumbnail_process":"video/snapshot,t_0,f_jpg,ar_auto,w_300","fields":"*","order_by":"updated_at","order_direction":"DESC"}`, drive_id, fileid))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Authorization", uploadData["access_token"].(string))
	req.Header.Add("Content-Type", "text/plain")

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return result
}

func userSelect() {
	fmt.Println("请选择目录或文件")
	var position string
	for i := 0; i > -1; i++ {
		if i == 0 {
			position = "root"
		}
		files := listFiles(position)["items"].([]interface{})
		for j := 0; j < len(files); j++ {
			fileType := files[j].(map[string]interface{})["type"].(string)
			if fileType == "folder" {
				fmt.Printf("\x1b[%dm%s \x1b[0m\n", 32, fmt.Sprintf("%d", j)+". "+files[j].(map[string]interface{})["name"].(string))
			} else {
				fmt.Println(fmt.Sprintf("%d", j) + ". " + files[j].(map[string]interface{})["name"].(string))
			}

		}
		var nextIndex int
		fmt.Printf("请输入您的选择:")
		fmt.Scanf("%d", &nextIndex)
		if nextIndex >= len(files) {
			log.Fatal("错误的序号")
		}
		fileItem := files[nextIndex].(map[string]interface{})
		if fileItem["type"].(string) == "folder" {
			position = fileItem["file_id"].(string)
		} else {
			startDownload(fileItem["file_id"].(string), fileItem["name"].(string))
			break
		}
		fmt.Println("")

	}
}

func startDownload(fileid string, filename string) {
	fmt.Printf("您要下载到哪里(./)(末尾加上/):")
	var targetPath string
	fmt.Scanf("%s", &targetPath)
	if targetPath == "" {
		targetPath = "./"
	}
	url := getDownloadUrl(fileid)

	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("referer", "https://www.aliyundrive.com/")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	f, err := os.Create(targetPath + filename)
	if err != nil {
		panic(err)
	}
	io.Copy(f, res.Body)
	fmt.Println("下载完成 - " + filename)

}

func getDownloadUrl(fileid string) string {
	var authKey string
	if programConfig["refreshToken"] == nil {
		if localConfig["refresh_token"] == nil {
			log.Fatal("Refresh Token未配置")
		} else {
			authKey = localConfig["refresh_token"].(string)
		}
	} else {
		authKey = programConfig["refreshToken"].(string)
	}
	uploadData := getUploadData(authKey)
	updateRefreshToken(uploadData["refresh_token"].(string))
	drive_id := uploadData["default_drive_id"].(string)
	url := "https://api.aliyundrive.com/v2/file/get_download_url"
	method := "POST"
	payload := strings.NewReader(fmt.Sprintf(`{"drive_id":"%s","file_id":"%s"}`, drive_id, fileid))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", uploadData["access_token"].(string))
	req.Header.Add("Content-Type", "text/plain")

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return result["url"].(string)

}

func downloadProcess() {
	if programConfig["downloadFileid"] == nil {
		userSelect()
	}
}

func server() {
	r := gin.Default()
	r.Use(Cors())
	r.POST("/getUpload", uploadFunc)
	r.POST("/directUpload", directUploadFunc)
	r.POST("/complete", completeFunc)
	runPort := programConfig["port"]
	if runPort == nil {
		runPort = "13142"
	} else {
		runPort = runPort.(string)
	}

	r.Run(fmt.Sprintf(":%s", runPort))
}

func main() {
	args := os.Args[1:]
	configCheck("config.json")
	programConfig = make(map[string]interface{})
	commandLinePreProcess(args)
	Process()
}
