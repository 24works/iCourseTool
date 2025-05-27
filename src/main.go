package main

import (
	ASNIColor "CourseTool/ASNIcolor"
	"CourseTool/sdtbu"
	"fmt"
	"log"
)

func main() {
	fmt.Println("Hello, World!")
	sdtbu.Init() // 初始化您的套件

	// 創建一個新的客戶端會話
	session, err := sdtbu.NewClientSession()
	if err != nil {
		log.Fatalf("Failed to create client session: %v", err)
	}

	// 使用這個會話進行登入
	username := ""
	password := ""
	err = session.Login(username, password)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	fmt.Println(ASNIColor.BrightWhite + ASNIColor.BgCyan + "Login successful!" + ASNIColor.Reset)

	// 現在您可以使用 `session` 對象在其他函數中執行需要登入狀態的操作
	// 例如：
	// session.FetchCourseList()
	// session.SubmitAssignment()

	session.GetClassbyUserInfo()
	session.GetClassbyTime()
	session.ParseClassList(session.ClassListbyTimeString)
}
