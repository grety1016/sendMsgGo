package ddtoken

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sendmsggo/util/mssql"
	"sync"
	"time"

	// "sendmsggo/util/eventful"

	// "sendmsggo/util/mssql/mssqldemo"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// DDToken 结构体(working)
type DDToken struct {
	URL       string `json:"url"`
	AppKey    string `json:"appkey"`
	AppSecret string `json:"appsecret"`
}

// DDTokenResult 结构体(working)
type DDTokenResult struct {
	ErrCode     int    `json:"errcode"`
	AccessToken string `json:"access_token"`
	ErrMsg      string `json:"errmsg"`
}

// 创建ddtoken请求实例(working)
func NewDDToken(url, appkey, appsecret string) *DDToken {
	return &DDToken{
		URL:       url,
		AppKey:    appkey,
		AppSecret: appsecret,
	}
}

// 获取钉钉机器人token方法(working)
func (d *DDToken) GetToken() (string, error) {
	// 将获取token参数加入到一个URL查询参数
	params := url.Values{}
	params.Add("appkey", d.AppKey)
	params.Add("appsecret", d.AppSecret)

	// 发起GET请求
	resp, err := http.Get(fmt.Sprintf("%s?%s", d.URL, params.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应主体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析JSON响应
	accessToken := gjson.GetBytes(body, "access_token")
	if !accessToken.Exists() {
		errMsg := gjson.GetBytes(body, "errmsg").String()
		return "", fmt.Errorf("failed to get access token: %s", errMsg)
	}

	return accessToken.String(), nil
}

// 通过useriphone获取userid(working)
type DDUserid struct{}

// 获取userid成功时usrid类型(working)
type DDUseridValue struct {
	UserID string `json:"userid"`
}

// 获取userid结果返回类型,包含错误类型（working）
type DDUseridResult struct {
	ErrCode   int           `json:"errcode"`
	Result    DDUseridValue `json:"result"`
	ErrMsg    string        `json:"errmsg"`
	RequestID string        `json:"request_id"`
}

// 钉钉DDUserid实现(isworking)
func (d *DDUserid) GetUserID(accessToken, mobile string) (string, error) {
	request := map[string]string{
		"access_token": accessToken,
		"mobile":       mobile,
	}
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://oapi.dingtalk.com/topapi/v2/user/getbymobile", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var useridResult DDUseridResult
	err = json.Unmarshal(body, &useridResult)
	if err != nil {
		return "", err
	}

	if useridResult.ErrCode != 0 {
		return "", fmt.Errorf("failed to get user id: %s", useridResult.ErrMsg)
	}

	return useridResult.Result.UserID, nil
}

// 发送短信验证码结构体
type SmsMessage struct {
	DDToken   string `json:"ddtoken"`
	DDUserid  string `json:"dduserid"`
	UserPhone string `json:"userphone"`
	RobotCode string `json:"robotcode"`
	SmsCode   int    `json:"smscode"`
}

func (s *SmsMessage) SendSmsCode() error {
	// 将字符串形式的数组转换为字符串数组
	var userIds []string

	//GO语言中json解析数组时，需要将字符串形式的数组转换为字符串数组，不能直接将字符串转换为字符串数组
	err := json.Unmarshal([]byte(s.DDUserid), &userIds)
	if err != nil {
		logrus.Errorf("Error parsing userIds: %v", err)
		return err
	}

	// 创建请求表头结构
	requestHeaders := map[string]string{
		"x-acs-dingtalk-access-token": s.DDToken,
	}

	// 创建请求表体结构
	msgParams := fmt.Sprintf(`{"msgtype": "text", "content": "%d"}`, s.SmsCode)
	requestBody := map[string]interface{}{
		"msgParam":  msgParams,
		"msgKey":    "sampleText",
		"robotCode": s.RobotCode,
		"userIds":   userIds, // 确保 userIds 是字符串数组
	}

	// 将请求体转换为 JSON
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		logrus.Errorf("Error marshaling JSON: %v", err)
		return err
	}
	logrus.Infof("Request Body JSON: %s", string(requestBodyJSON))

	// 发起请求
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		logrus.Errorf("Error creating request: %v", err)
		return err
	}

	// 设置请求头
	for key, value := range requestHeaders {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求并读取响应
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("Error sending request: %v", err)
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Error reading response body: %v", err)
		return err
	}

	logrus.Infof("Send SMS code result: %s, userphone: %s", string(respBody), s.UserPhone)
	return nil
}

// 审批流程结构体
type ProcessApproval struct {
	AccessToken string
	MsgParams   string
	MsgKey      string
	RobotCode   string
	UserID      string
}

// SendMsg 发送消息到当前用户钉钉账号
func (p *ProcessApproval) SendMsg() error {
	// 创建请求表头结构
	requestHeaders := map[string]string{
		"x-acs-dingtalk-access-token": p.AccessToken,
	}

	// 创建请求表体结构
	requestBody := map[string]interface{}{
		"msgParam":  p.MsgParams,
		"msgKey":    p.MsgKey,
		"robotCode": p.RobotCode,
		"userIds":   p.UserID,
	}

	// 将请求体转换为 JSON
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	// 发起请求
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return err
	}

	// 设置请求头
	for key, value := range requestHeaders {
		req.Header.Set(key, value)
	}

	// 发送请求并读取响应
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	logrus.Infof("send_result: %s, userphone: %s\n", string(respBody), p.UserID)
	return nil
}

// Userid 结构体(working)
type Userid struct {
	Username  string
	Userphone string
	DDUserid  *string
}

// NewUserid 创建新的 Userid 实例(working)
func NewUserid(username, userphone string) *Userid {
	return &Userid{
		Username:  username,
		Userphone: userphone,
		DDUserid:  nil,
	}
}

// SendMSG 结构体(working)
type SendMSG struct{}

// NewSendMSG 创建新的 SendMSG 实例(working)
func NewSendMSG() *SendMSG {
	return &SendMSG{}
}

type DBConfig = mssql.DBConfig
type DB = mssql.DBWrapper

var dbonce sync.Once

// 初始化数据库连接池(working)
func (smg *SendMSG) initDB() (*DB, error) {
	var initErr error
	var db *DB
	dbonce.Do(func() {
		var dbconfigstr = "server=47.103.31.8;port=1433;user id=kxs_dev;password=kephi;database=Kxs_Interface;encrypt=true;trustServerCertificate=true;connection timeout=30;application name=sendmsg"
		dbConfig := mssql.SetDBConfig(dbconfigstr, 4, 2, 60*time.Minute, 15*time.Minute)
		db, initErr = mssql.InitDB(dbConfig)
		if initErr != nil {
			logrus.Fatalf("[DD] - Failed to open to database: %v", initErr)
			return
		}
	})
	return db, initErr
}

// GetSendNum 查询当前待办流程需要发送消息的行数
func (s *SendMSG) GetSendNum(db *sql.DB) (int, error) {
	var num int
	err := db.QueryRow("CALL get_flow_list()").Scan(&num)
	if err != nil {
		return 0, err
	}
	return num, nil
}

// GetUserlist 获取userid表中未有userid的用户并回写userid(working)
func (s *SendMSG) GetUserListUserid(db *DB, gzymAccessToken, zbAccessToken string) error {
	var useridList []*Userid
	if err := db.QueryCollect(&useridList, "SELECT username, rtrim(ltrim(userphone)) AS userphone, rtrim(ltrim(ddUserID)) AS dduserid FROM sendMsg_users WHERE ISNULL(ddUserID, '') = ''"); err != nil {
		return err
	}

	for _, user := range useridList {

		userid, err := s.getDDUserID(gzymAccessToken, user.Userphone)
		if err != nil {
			return err
		}
		if userid != "" {
			userparam := map[string]interface{}{"userid": fmt.Sprintf("[\"%s\"]", userid), "userphone": user.Userphone}

			_, err := db.ExecSQLWithTran("UPDATE sendMsg_users SET dduserid = :userid, ddtoken = 'gzym_access_token', robotcode = 'dingrw2omtorwpetxqop' WHERE userphone = :userphone", userparam)
			if err != nil {
				return err
			}
		} else {
			userid, err = s.getDDUserID(zbAccessToken, user.Userphone)
			if err != nil {
				return err
			}
			userparam := map[string]interface{}{"userid": fmt.Sprintf("[\"%s\"]", userid), "userphone": user.Userphone}

			_, err := db.ExecSQLWithTran("UPDATE sendMsg_users SET dduserid = :userid, ddtoken = 'zb_access_token', robotcode = 'dingzblrl7qs6pkygqcn' WHERE userphone = :userphone", userparam)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// getDDUserID 调用通过手机获取userid(working)
func (s *SendMSG) getDDUserID(accessToken, mobile string) (string, error) {

	// 创建请求 URL，并将 access_token 作为查询参数
	url := fmt.Sprintf("https://oapi.dingtalk.com/topapi/v2/user/getbymobile?access_token=%s", accessToken)

	// 创建请求体
	requestBody := map[string]string{
		"mobile": mobile,
	}
	requestBodyJSON, _ := json.Marshal(requestBody)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBodyJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		logrus.Errorf("[DD] - Get dduserid failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("[DD] - Read dduserid failed: %v", err)
	}

	//创建响应结构体
	var useridResult struct {
		ErrCode int    `json:"errcode"`
		SubCode string `json:"sub_code"`
		SubMsg  string `json:"sub_msg"`
		ErrMsg  string `json:"errmsg"`
		Result  struct {
			UserID string `json:"userid"`
		} `json:"result"`
	}
	err = json.Unmarshal(respBody, &useridResult)
	if err != nil {
		logrus.Errorf("[DD] - Json unmarshal dduserid failed: %v", err)
	}
	if useridResult.ErrCode == 88 {
		return "", fmt.Errorf("[DD] - Get dduserid failed:%v", useridResult)
	}
	return useridResult.Result.UserID, nil
}

// LocalThread 本地线程函数,查询数据库用户列表是否有新的用户加入，获取dduserid并回写到数据库中(working)
func LocalThread() error {
	sendmsg := NewSendMSG() // 创建消息发送对象

	// 获取一个数据连接池对象
	db, err := sendmsg.initDB()
	if err != nil {
		logrus.Fatalf("[DD] @%s - Failed to open to database: %v }", db.DBName(), err)
		return err
	}
	logrus.Infof("[DD] @%s - Connecte database success ", db.DBName())

	defer db.Close()

	// 先获取当前消息用户列表中是否存在没有DDuserid的用户，有的话查询出来并从钉钉接口中获取userid
	var addNewUsers int64
	if row, err := db.QueryValue("DECLARE @row INT; EXEC CheckForNewAddedUsers @row OUTPUT; SELECT @row", nil); err != nil {
		if err == sql.ErrNoRows {
			addNewUsers = 0
		} else {
			logrus.Errorf("[DD]-执行CheckForNewAddedUsers过程时出错: %v\n", err)
		}
	} else {
		if num, ok := row.(int64); ok {
			addNewUsers = num
		}
	}

	if addNewUsers > 0 {
		logrus.Infof("[DD]-用户列表中dduserid为空的用户数: %d", addNewUsers)

		// 初始化广州野马获取access_token的对象
		gzymDDToken := NewDDToken(
			"https://oapi.dingtalk.com/gettoken",
			"dingrw2omtorwpetxqop",
			"Bcrn5u6p5pQg7RvLDuCP71VjIF4ZxuEBEO6kMiwZMKXXZ5AxQl_I_9iJD0u4EQ-N",
		)

		// 获取广州野马实时access_token
		gzymAccessToken, err := gzymDDToken.GetToken()
		if err != nil {
			logrus.Errorf("[DD]-Failed to get Guangzhou Yema Token: %v", err)
			return err

		}
		logrus.Infof("[DD]-广州野马Token: %s, robotcode: dingrw2omtorwpetxqop", gzymAccessToken)

		// 初始化总部获取access_token的对象
		zbDDToken := NewDDToken(
			"https://oapi.dingtalk.com/gettoken",
			"dingzblrl7qs6pkygqcn",
			"26GGYRR_UD1VpHxDBYVixYvxbPGDBsY5lUB8DcRqpSgO4zZax427woZTmmODX4oU",
		)

		// 获取总部实时access_token
		zbAccessToken, err := zbDDToken.GetToken()
		if err != nil {
			logrus.Errorf("[DD]-Failed to get Zongbu Token: %v", err)
			return err
		}
		logrus.Infof("[DD]-福建快先森Token: %s, robotcode: dingzblrl7qs6pkygqcn", zbAccessToken)

		// 循环遍历用户列表中未有dduserid的用户,并回写到消息用户列表中
		err = sendmsg.GetUserListUserid(db, gzymAccessToken, zbAccessToken)
		if err != nil {
			logrus.Errorf("[DD]-Failed to get user list: %v", err)
			return err
		}
	}
	return nil
}
