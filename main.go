package main

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"

	"todolist.go/db"
	"todolist.go/service"
)

const port = 8000

func Floor(f float64) int {
    return int(f)
}

func main() {
	// initialize DB connection
	dsn := db.DefaultDSN(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	if err := db.Connect(dsn); err != nil {
		log.Fatal(err)
	}

	// initialize Gin engine
	engine := gin.Default()
	engine.SetFuncMap(template.FuncMap{
		"Floor": Floor,
	})
	engine.LoadHTMLGlob("views/*.html")
	// prepare session
    store := cookie.NewStore([]byte("my-secret"))
    engine.Use(sessions.Sessions("user-session", store))

	// routing
	engine.Static("/assets", "./assets")
	engine.GET("/", service.Home)
	engine.GET("/list", service.LoginCheck, service.TaskList)
	engine.GET("/task/new", service.LoginCheck, service.NewTaskForm)
	engine.POST("/task/new", service.LoginCheck, service.RegisterTask)
	taskGroup := engine.Group("/task")
    taskGroup.Use(service.LoginCheck, service.TaskCheck)
    {
        taskGroup.GET("/:id", service.ShowTask)
        taskGroup.GET("/edit/:id", service.EditTaskForm)
        taskGroup.POST("/edit/:id", service.UpdateTask)
        taskGroup.GET("/delete/:id", service.DeleteTask)
    }
	// ユーザ登録
    engine.GET("/user/new", service.NewUserForm)
    engine.POST("/user/new", service.RegisterUser)
	// ユーザー情報変更
	engine.GET("/user/edit", service.LoginCheck, service.EditUserForm)
    engine.POST("/user/edit", service.LoginCheck, service.ReregisterUser)
	// ログイン処理
	engine.GET("/login", service.LoginUserForm)
	engine.POST("/login", service.Login)
	// ログアウト処理
	engine.GET("/logout", service.Logout)
	// アカウント削除処理
	engine.GET("/user/delete", service.LoginCheck, service.DeleteUser)

	// start server
	engine.Run(fmt.Sprintf(":%d", port))
}
