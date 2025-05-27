package sdtbu

import (
	"CourseTool/des" // 假設 des 套件用於加密
	"bytes"
	"encoding/json" // 導入 json 套件，用於處理 JSON 數據
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url" // 導入 url 套件，用於構建表單數據
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// ANSI 顏色代碼常量
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Purple = "\033[35m" // 新增紫色
)

// Init 函數，用於初始化
func Init() {
	fmt.Println(Green + "SDTBU: Initializing..." + Reset)
}

// LoginParams 結構體用於儲存從登入頁面提取的參數
type LoginParams struct {
	Lt        string
	Execution string
	EventId   string
}

// ClientSession 結構體用於儲存 HTTP 客戶端和 cookie jar，以便在不同函數間共用
type ClientSession struct {
	Client *http.Client
	Jar    *cookiejar.Jar
	// 您可以在這裡添加其他會話相關的資訊，例如 User-Agent
	UserAgent string
	reqURL    string // 用於存儲請求的 URL
	// 您可以在這裡添加其他需要的字段，例如請求后獲得的部分信息
	CalssListUserInfoString string // 用於存儲課程列表的字符串
	ClassListbyTimeString   string // 用於存儲本周課程時間列表的字符串
}

// NewClientSession 函數用於創建並初始化一個新的 ClientSession
func NewClientSession() (*ClientSession, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf(Red+"SDTBU: Failed to create cookie jar: %v"+Reset, err)
	}

	client := &http.Client{
		Jar: jar, // 為客戶端設定 cookie jar
	}

	return &ClientSession{
		Client:    client,
		Jar:       jar,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0",
	}, nil
}

// GetClassbyTime 函數用於發送 POST 請求獲取用戶的本周課程資訊
func (cs *ClientSession) GetClassbyTime() error {
	fmt.Println(Blue + "SDTBU: Fetching class information by time..." + Reset)

	// 請求 URL
	var requestURL string

	//判斷是否使用webVPN
	if cs.reqURL != "https://zhss.sdtbu.edu.cn/tp_up/" {
		requestURL = cs.reqURL + "up/widgets/getClassbyTime?vpn-12-o2-zhss.sdtbu.edu.cn"
	} else {
		requestURL = cs.reqURL + "up/widgets/getClassbyTime"
	}

	// 聲明一個 Go 切片變量，用於存儲解析後的 classList 內容
	// 這裡我們將 classList 內的物件鍵值改為 interface{}，以適應可能包含數字或其他類型的 JSON 值
	var classListContent []map[string]interface{}

	// 使用 json.Unmarshal 將字符串變量解析到 Go 切片中
	err := json.Unmarshal([]byte(cs.CalssListUserInfoString), &classListContent)
	if err != nil {
		fmt.Println(Red+"Error unmarshalling classList string:", err, Reset)
		return fmt.Errorf(Red+"SDTBU: Error unmarshalling classListUserInfoString: %v"+Reset, err)
	}

	// 構建請求體數據
	requestBody := map[string]interface{}{
		"schoolYear": "2024-2025",
		"semester":   "2",
		"learnWeek":  "17",
		"classList":  classListContent, // 使用之前獲取的課程列表
	}

	// 將請求體數據編碼為 JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error marshalling request body to JSON: %v"+Reset, err)
	}

	// 創建 POST 請求
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error creating POST request for GetClassbyTime: %v"+Reset, err)
	}

	// 設定請求標頭
	req.Header.Set("Content-Type", "application/json") // 設定內容類型為 JSON
	req.Header.Set("User-Agent", cs.UserAgent)         // 設定 User-Agent

	// 發送請求
	resp, err := cs.Client.Do(req)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error sending POST request to GetClassbyTime: %v"+Reset, err)
	}
	defer resp.Body.Close() // 確保響應主體已關閉

	fmt.Printf(Cyan+"SDTBU: POST request to %s status: %s\n"+Reset, requestURL, resp.Status)

	// 讀取響應主體
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error reading GetClassbyTime response body: %v"+Reset, err)
	}

	// 記錄Class内容
	cs.ClassListbyTimeString = string(bodyBytes)

	// 打印響應內容
	fmt.Println(White + "SDTBU: GetClassbyTime response content:" + Reset)
	fmt.Println(string(bodyBytes))

	// 您可以在這裡進一步解析 JSON 響應，例如：
	// var classInfo map[string]interface{}
	// if err := json.Unmarshal(bodyBytes, &classInfo); err != nil {
	// 	return fmt.Errorf("SDTBU: Error unmarshalling class info JSON: %v", err)
	// }
	// fmt.Printf("SDTBU: Parsed class info: %+v\n", classInfo)

	return nil
}

// GetClassbyUserInfo 函數用於發送 POST 請求獲取用戶的課程資訊
func (cs *ClientSession) GetClassbyUserInfo() error {
	fmt.Println(Blue + "SDTBU: Fetching class information..." + Reset)

	// 請求 URL
	var requestURL string

	//判斷是否使用webVPN
	if cs.reqURL != "https://zhss.sdtbu.edu.cn/tp_up/" {
		requestURL = cs.reqURL + "up/widgets/getClassbyUserInfo?vpn-12-o2-zhss.sdtbu.edu.cn"
	} else {
		requestURL = cs.reqURL + "up/widgets/getClassbyUserInfo"
	}

	// 構建請求體數據
	requestBody := map[string]string{
		"schoolYear": "2024-2025",
		"semester":   "2",
		"learnWeek":  "14",
	}

	// 將請求體數據編碼為 JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error marshalling request body to JSON: %v"+Reset, err)
	}

	// 創建 POST 請求
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error creating POST request for getClassbyUserInfo: %v"+Reset, err)
	}

	// 設定請求標頭
	req.Header.Set("Content-Type", "application/json") // 設定內容類型為 JSON
	req.Header.Set("User-Agent", cs.UserAgent)         // 設定 User-Agent

	// 發送請求
	resp, err := cs.Client.Do(req)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error sending POST request to getClassbyUserInfo: %v"+Reset, err)
	}
	defer resp.Body.Close() // 確保響應主體已關閉

	fmt.Printf(Cyan+"SDTBU: POST request to %s status: %s\n"+Reset, requestURL, resp.Status)

	// 讀取響應主體
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error reading getClassbyUserInfo response body: %v"+Reset, err)
	}

	// 記錄Class内容
	cs.CalssListUserInfoString = string(bodyBytes)

	// 打印響應內容
	// fmt.Println("SDTBU: GetClassbyUserInfo response content:")
	// fmt.Println(string(bodyBytes))

	// 您可以在這裡進一步解析 JSON 響應，例如：
	// var classInfo map[string]interface{}
	// if err := json.Unmarshal(bodyBytes, &classInfo); err != nil {
	// 	return fmt.Errorf("SDTBU: Error unmarshalling class info JSON: %v", err)
	// }
	// fmt.Printf("SDTBU: Parsed class info: %+v\n", classInfo)

	return nil
}

// Login 函數，作為 ClientSession 的方法，用於登入 SDTBU 系統
// 該函數模擬了瀏覽器行為，先進行 GET 請求獲取登入頁面，
// 然後解析頁面以提取必要的參數（例如 lt, execution, _eventId 值），
// 最後構建 POST 請求並發送登入資訊。
func (cs *ClientSession) Login(username, password string) error {
	fmt.Printf(Green+"SDTBU: Logging in with username: %s\n"+Reset, username)

	// --- 1. 執行 GET 請求以獲取登入頁面和相關參數 ---
	// 宣告 req 變數，以便在後續的 GET 和 POST 請求中重複使用
	var req *http.Request
	var err error
	var resp *http.Response

	getReqURL := "https://zhss.sdtbu.edu.cn/tp_up/"
	req, err = http.NewRequest("GET", getReqURL, nil)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error creating GET request: %v"+Reset, err)
	}
	req.Header.Set("User-Agent", cs.UserAgent)

	// 使用共用的客戶端傳送 GET 請求
	resp, err = cs.Client.Do(req)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error making GET request to %s: %v"+Reset, getReqURL, err)
	}
	defer resp.Body.Close() // 確保 GET 請求的響應主體已關閉

	fmt.Printf(Cyan+"SDTBU: GET request to %s status: %s\n"+Reset, getReqURL, resp.Status)

	// 讀取響應主體以提取登入表單的 HTML 內容
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error reading GET response body: %v"+Reset, err)
	}
	htmlBody := string(bodyBytes)

	// POST 請求的 URL 將是提供登入表單的 URL。
	// 在 GET 請求（以及任何重定向）之後，這在 resp.Request.URL 中可用。
	postTargetURL := resp.Request.URL.String()
	fmt.Println(Yellow+"SDTBU: Login form URL (target for POST): "+Reset, postTargetURL)

	// 2. 從 HTML 內容中提取登入參數 (lt, execution, _eventId)
	// 這些參數通常是隱藏欄位，用於維持會話狀態或防止 CSRF 攻擊。
	loginParams := ExtractLoginParameters(htmlBody)
	if loginParams.Lt == "" || loginParams.Execution == "" || loginParams.EventId == "" {
		return fmt.Errorf(Red+"SDTBU: Failed to extract all required login parameters. Lt: '%s', Execution: '%s', EventId: '%s'"+Reset,
			loginParams.Lt, loginParams.Execution, loginParams.EventId)
	}
	fmt.Println(Yellow+"SDTBU: Extracted login parameters:", loginParams, Reset)

	// 3. 準備 POST 請求的表單資料
	// 根據原代碼邏輯，rsa 值由用戶名、密碼和 lt 值拼接而成，然後進行加密。
	rsa := fmt.Sprintf("%s%s%s", username, password, loginParams.Lt)
	//fmt.Println("SDTBU: Prepared RSA value (before encryption):", rsa)

	// 加密 RSA 值，這裡假設 des.StrEnc 函數可用於加密
	encryptedRSA := des.StrEnc(rsa, "1", "2", "3")
	//fmt.Println("SDTBU: Encrypted RSA value:", encryptedRSA)

	// 計算用戶名和密碼的長度，用於 POST 請求中的 ul 和 pl 參數
	ul := utf8.RuneCountInString(username)
	pl := utf8.RuneCountInString(password)

	// 構建 POST 請求的資料，使用 url.Values 更規範地處理表單數據
	formData := url.Values{}
	formData.Set("rsa", encryptedRSA)
	formData.Set("ul", fmt.Sprintf("%d", ul))
	formData.Set("pl", fmt.Sprintf("%d", pl))
	formData.Set("lt", loginParams.Lt)
	formData.Set("execution", loginParams.Execution)
	formData.Set("_eventId", loginParams.EventId)

	postDataString := formData.Encode() // 將表單數據編碼為 URL 查詢字符串格式
	//fmt.Println("SDTBU: POST data:", postDataString)

	// 將 POST 資料字串轉換為 io.Reader
	postDataReader := strings.NewReader(postDataString)

	// --- 4. 執行 POST 請求以提交登入資訊 ---
	// 這裡重新賦值 req，而不是重新宣告
	req, err = http.NewRequest("POST", postTargetURL, postDataReader)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error creating POST request: %v"+Reset, err)
	}

	// 設定請求標頭
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", cs.UserAgent) // 保持 User-Agent 一致

	// 傳送 POST 請求
	resp, err = cs.Client.Do(req)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error sending POST request: %v"+Reset, err)
	}
	defer resp.Body.Close() // 確保 POST 響應主體已關閉

	fmt.Printf(Cyan+"SDTBU: POST request to %s status: %s\n"+Reset, req.URL, resp.Status)
	fmt.Println(Yellow+"SDTBU: Current URL after POST:", resp.Request.URL.String()+Reset) // 列印請求的最終 URL

	cs.reqURL = resp.Request.URL.String() // 儲存最終請求的 URL

	// 5. 打印 Cookies (可選，用於調試)
	// cookies := cs.Client.Jar.Cookies(resp.Request.URL) // 使用響應中的最終 URL 來獲取 cookies
	// if len(cookies) > 0 {
	// 	fmt.Println("SDTBU: Cookies after POST request:")
	// 	for _, cookie := range cookies {
	// 		fmt.Printf("  Name: %s, Value: %s, Domain: %s, Path: %s, Expires: %s, HttpOnly: %t, Secure: %t\n",
	// 			cookie.Name, cookie.Value, cookie.Domain, cookie.Path, cookie.Expires, cookie.HttpOnly, cookie.Secure)
	// 	}
	// } else {
	// 	fmt.Println("SDTBU: No cookies found after POST request for URL:", resp.Request.URL.String())
	// }

	// 在這裡可以添加進一步的邏輯來檢查登入是否成功，
	// 例如：檢查響應狀態碼、重定向的 URL、或響應主體內容。
	// 如果登入成功，通常會重定向到一個受保護的頁面。

	// 請求用戶主頁或儀表板頁面
	// 這裡重新賦值 req，而不是重新宣告
	req, err = http.NewRequest("GET", resp.Request.URL.String()+"view?m=up", nil)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error creating GET request for dashboard: %v"+Reset, err)
	}
	req.Header.Set("User-Agent", cs.UserAgent) // 保持 User-Agent 一致
	resp, err = cs.Client.Do(req)
	if err != nil {
		return fmt.Errorf(Red+"SDTBU: Error fetching dashboard page: %v"+Reset, err)
	}
	defer resp.Body.Close()

	// 打印主頁或儀表板頁面
	// bodyBytes, err = io.ReadAll(resp.Body)
	// if err != nil {
	// 	return fmt.Errorf("SDTBU: Error reading dashboard response body: %v", err)
	// }
	// fmt.Println("SDTBU: Dashboard page content:")
	// fmt.Println(string(bodyBytes)) // 打印主頁內容，這裡可以進一步解析或處理
	return nil
}

// ExtractLoginParameters 從 HTML 內容中提取登入參數 (lt, execution, _eventId) 的值
// 該函數會遍歷 HTML 節點，查找具有特定 id 或 name 屬性的 input 標籤，
// 並提取其 value 值。
func ExtractLoginParameters(htmlbody string) LoginParams {
	doc, err := html.Parse(strings.NewReader(htmlbody))
	if err != nil {
		log.Printf(Red+"SDTBU: Error parsing HTML: %v"+Reset, err)
		return LoginParams{} // 返回一個空的 LoginParams
	}

	params := LoginParams{}
	var findParams func(*html.Node)
	findParams = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			var id, name, value string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "id":
					id = attr.Val
				case "name":
					name = attr.Val
				case "value":
					value = attr.Val
				}
			}

			// 根據 id 或 name 提取對應的參數值
			if id == "lt" {
				params.Lt = value
			}
			if name == "execution" {
				params.Execution = value
			}
			if name == "_eventId" {
				params.EventId = value
			}
		}
		// 遞歸遍歷子節點
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findParams(c)
			// 如果所有參數都已找到，則停止遍歷，提高效率
			if params.Lt != "" && params.Execution != "" && params.EventId != "" {
				return
			}
		}
	}

	findParams(doc)
	return params
}
