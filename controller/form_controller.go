package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sendmsggo/cstvalidator"
	"sendmsggo/model"
	"sendmsggo/util/logger"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetItemList 待办列表
func GetItemList(c *gin.Context) {
	var msg string //返回信息
	var code int   //返回状态码

	var getItemList model.GetItemList

	err := c.ShouldBindQuery(&getItemList)
	if err != nil {
		msg = "请求参数不正确"
		code = http.StatusBadRequest
		ResponseSuccess(c, code, msg, "", 1)
		return
	}

	//从validator自定义验证器中获取验证器实例调用正则验证手机号及验证码格式
	err = cstvalidator.GetValidator().Struct(getItemList) //验证结构体
	if err != nil {
		msg = "手机号或待办状态码不正确!"
		code = http.StatusBadRequest
		ResponseSuccess(c, code, msg, "", 1)
		return
	}

	//查询数据库获取数据
	db := c.MustGet("db").(*DB)

	var itemList []model.FlowItemList = make([]model.FlowItemList, 0) //待办列表接收对象

	params := []interface{}{sql.Named("itemstatus", getItemList.ItemStatus), sql.Named("userphone", getItemList.UserPhone)} //参数化查询

	query := "SELECT * FROM getTodoList(@itemstatus,@userphone)" //查询语句

	err = db.QueryCollect(&itemList, query, params...) //查询执行结果
	if err != nil {
		msg = "获取待办列表失败!"
		code = http.StatusInternalServerError
		ResponseSuccess(c, code, msg, "", 1)
		return
	} else {
		ResponseSuccess(c, http.StatusOK, "获取成功", itemList, len(itemList))
	}

}

// 获取用户流程明细路由（费用报销与差旅报销）
func GetItemDetailFybxAndClbx(c *gin.Context) {
	var msg string //返回信息
	var code int   //返回状态码
	var flowdetailfybxandclbx []model.FlowItemDetailFybxAndClbx = make([]model.FlowItemDetailFybxAndClbx, 0)

	var fProcinstID = c.Query("fprocinstid")
	var fPhone = c.MustGet("fphone")

	db := c.MustGet("db").(*DB)

	//查询数据库获取数据
	query := "SELECT * FROM getFlowDetailFybxAndClbx(@fprocinstID,@fPhone)"                     //查询语句
	params := []interface{}{sql.Named("fprocinstid", fProcinstID), sql.Named("fphone", fPhone)} //参数化查询
	err := db.QueryCollect(&flowdetailfybxandclbx, query, params...)
	if err != nil {
		msg = "获取待办明细失败!"
		code = http.StatusInternalServerError
		ResponseSuccess(c, code, msg, "", 1)
		return
	} else {
		ResponseSuccess(c, http.StatusOK, "获取成功", flowdetailfybxandclbx, len(flowdetailfybxandclbx))
	}

}

type Attachment = model.Attachment // 附件类型别名

// 获取用户流程明细报销明细路由（费用报销）
func GetFlowDetailRowsFybx(c *gin.Context) {
	var msg string //返回信息
	var code int   //返回状态码
	var flowItemDetailRowFybx []model.FlowDetailRowFybx = make([]model.FlowDetailRowFybx, 0)

	var fProcinstID = c.Query("fprocinstid")

	db := c.MustGet("db").(*DB)

	//查询数据库获取数据
	query := "SELECT * FROM getFlowDetailRowsFybx(@fprocinstID)" //查询语句
	err := db.QueryCollect(&flowItemDetailRowFybx, query, sql.Named("fprocinstid", fProcinstID))
	if err != nil {
		msg = "获取待办明细行失败!"
		code = http.StatusInternalServerError
		ResponseSuccess(c, code, msg, "", 1)
		return
	}

	for i := range flowItemDetailRowFybx {
		if flowItemDetailRowFybx[i].FSnnaAttachments == "" {
			flowItemDetailRowFybx[i].FSnnaAttachments = "[]" // 初始化为空 JSON 数组
		}

		// 解析 JSON 到 Attachments 切片
		if err := json.Unmarshal([]byte(flowItemDetailRowFybx[i].FSnnaAttachments), &flowItemDetailRowFybx[i].Attachments); err != nil {
			flowItemDetailRowFybx[i].Attachments = []Attachment{}
		}

		flowItemDetailRowFybx[i].FSnnaAttachments = "" // 清空附件字符串

		handleAttachments(flowItemDetailRowFybx[i].Attachments, flowItemDetailRowFybx[i].Years) //处理附件数据

	}

	ResponseSuccess(c, http.StatusOK, "获取成功", flowItemDetailRowFybx, len(flowItemDetailRowFybx))
}

// 获取用户流程明细流程图路由
func GetFlowDetailFlowChart(c *gin.Context) {
	var msg string //返回信息
	var code int   //返回状态码
	var flowDetailFlowChart []model.FlowDetailFlowChart = make([]model.FlowDetailFlowChart, 0)

	var fProcinstID = c.Query("fprocinstid")

	db := c.MustGet("db").(*DB)

	//查询数据库获取数据
	query := "SELECT * FROM getFlowDetailChart(@fprocinstID)" //查询语句
	err := db.QueryCollect(&flowDetailFlowChart, query, sql.Named("fprocinstid", fProcinstID))
	if err != nil {
		msg = "获取流程图失败!"
		code = http.StatusInternalServerError
		ResponseSuccess(c, code, msg, "", 1)
		return
	} else {
		ResponseSuccess(c, http.StatusOK, "获取成功", flowDetailFlowChart, len(flowDetailFlowChart))
	}
}

// 查检附件是否存在
func CheckFileExist(c *gin.Context) {
	// var msg string //返回信息
	// var code int   //返回状态码

	var filePath = c.Query("filepath") //获取文件路径

	newPath := strings.Replace(filePath, "/sendmsg/files", "D:/kingdee  File", 1) //将网络路径转换为本地路径
	exist := fileExists(newPath)                                                  //检查新版本的文件是否存在

	if exist {
		ResponseSuccess(c, http.StatusOK, "文件存在", true, 1) //直接返回，说明可以直接访问
	} else {
		originPath, outerPath, fileExtNew := SplitPath(newPath) //解析路径，尝试获取旧版本的文件
		if exist = fileExists(originPath); exist {              //判断旧版本的文件是否存在
			success := turnFiles(originPath, outerPath, fileExtNew) //再次调用系统指令将相应的文档转换成较新版本
			if success {
				ResponseSuccess(c, http.StatusOK, "文件存在", true, 1) //转换成功
			} else {
				ResponseSuccess(c, http.StatusOK, "文件不存在", false, 1) //转换失败
			}
		} else {
			ResponseSuccess(c, http.StatusOK, "文件不存在", false, 1) //返回文件不存在
		}
	}
}

// #region 附件处理

// 处理附件数据
func handleAttachments(flowDetailRow []Attachment, year string) {
	for i := range flowDetailRow {
		item := &flowDetailRow[i]
		// 获取文件扩展名
		fileExt := strings.TrimPrefix(strings.ToLower(filepath.Ext(item.FileName)), ".")
		fileExtNew := fmt.Sprintf("%sx", fileExt)

		// 旧文件类型扩展名赋于新的扩展名
		if fileExt == "xls" || fileExt == "doc" || fileExt == "ppt" {
			item.FileType = fileExtNew
		} else {
			item.FileType = fileExt
		}

		// 按文件类型将文件名及路径拼接
		var filepath string
		switch fileExt {
		case "jpg", "png", "jpeg", "gif":
			filepath = fmt.Sprintf("/sendmsg/files/Image/%s/%s", year, item.ServerFileName)
		case "pdf", "docx", "xlsx", "pptx", "doc", "xls", "ppt":
			filepath = fmt.Sprintf("/sendmsg/files/Doc/%s/%s", year, item.ServerFileName)
		default:
			filepath = fmt.Sprintf("/sendmsg/files/Other/%s/%s", year, item.ServerFileName)
		}

		// 处理文件转换任务
		if fileExt == "xls" || fileExt == "doc" || fileExt == "ppt" {
			originPathCheck := fmt.Sprintf("D:\\kingdee  File\\doc\\%s\\%s.%s", year, item.ServerFileName, fileExt)
			buildPathCheck := fmt.Sprintf("D:\\kingdee  File\\doc\\%s\\%s.%s", year, item.ServerFileName, fileExtNew)

			if fileExists(originPathCheck) && !fileExists(buildPathCheck) {
				sourceFile := fmt.Sprintf("D:\\kingdee  File\\doc\\%s\\%s.%s", year, item.ServerFileName, fileExt)
				outerPath := fmt.Sprintf("D:\\kingdee  File\\doc\\%s", year)
				fmt.Println(sourceFile, outerPath, fileExtNew)
				go turnFiles(sourceFile, outerPath, fileExtNew) //调用系统指令将相应的文档转换成较新版本
			}
		}

		// 更新文件路径及名称
		if fileExt == "xls" || fileExt == "doc" || fileExt == "ppt" {
			item.ServerFileName = fmt.Sprintf("%s.%s", filepath, fileExtNew)
		} else {
			item.ServerFileName = fmt.Sprintf("%s.%s", filepath, fileExt)
		}

		// 计算文件大小
		if item.FileBytesLength/1024 >= 1024 {
			mbSize := item.FileBytesLength / 1024 / 1024
			size := fmt.Sprintf("%.2fMB", mbSize)
			item.FileSize = size
			item.FileBytesLength = 0
			item.FileLength = 0
		} else {
			kbSize := item.FileBytesLength / 1024
			size := fmt.Sprintf("%.2fKB", kbSize)
			item.FileSize = size
			item.FileBytesLength = 0
			item.FileLength = 0
		}
	}
}

// 判断指定目录文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// 调用系统指令将相应的文档转换成较新版本,返回布尔值（true：成功，false：失败）
func turnFiles(sourceFile, outerPath, fileExtClone string) bool {
	logger := logger.InitHTTPLogger() // 初始化日志记录器

	cmd := exec.Command("C:\\Program Files\\LibreOffice\\program\\soffice",
		"--headless", "--convert-to", fileExtClone, "--outdir", outerPath, sourceFile)
	logger.Infof("请求转换文件: %s", sourceFile)

	_, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("文件转换失败: %v", err)
		return false
	}

	logger.Infof("任务转换成功: %t , 转换结束", cmd.ProcessState.Success())
	return true
}

// SplitPath 函数将路径分成三部分返回
func SplitPath(path string) (string, string, string) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	filename := strings.TrimSuffix(base, ext)

	var newExt string
	// 去掉扩展名的最后一个字符
	if len(ext) > 1 {
		newExt = ext[:len(ext)-1] // 修改扩展名，去掉最后一个字符
		ext = ext[1:]
	}

	newPath := filepath.Join(dir, filename+newExt)

	return newPath, dir, ext
}

// #endregion 附件处理
