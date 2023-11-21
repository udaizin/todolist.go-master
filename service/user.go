package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

func NewUserForm(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "new_user_form.html", gin.H{"Title": "Register user"})
}

func hash(pw string) []byte {
    const salt = "todolist.go#"
    h := sha256.New()
    h.Write([]byte(salt))
    h.Write([]byte(pw))
    return h.Sum(nil)
}

func RegisterUser(ctx *gin.Context) {
    // フォームデータの受け取り
    username := ctx.PostForm("username")
    password := ctx.PostForm("password")
	password_confirmation := ctx.PostForm("password_confirmation")
    switch {
    case username == "":
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Usernane is not provided", "Username": username})
    case password == "":
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Password": password})
	case password_confirmation == "":
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password confirmation is not provided", "Password_confirmation": password_confirmation})
    }
    
    // DB 接続
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
	// パスワードが適切かチェック
	if utf8.RuneCountInString(password) <= 5 {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password must be at least 6 characters.", "Username": username, "Password": password, "Password_confirmation": password_confirmation})
		return
	} 
	_, err = strconv.Atoi(password)
	if err == nil {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Do not use only numbers for password.", "Username": username, "Password": password, "Password_confirmation": password_confirmation})
		return
	}
	
    // 重複チェック
    var duplicate int
    err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    if duplicate > 0 {
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username, "Password": password, "Password_confirmation": password_confirmation})
        return
    }

	// passwordとpassword_confirmationが同じかチェック
	if password != password_confirmation {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password and Password confirmarion is not equal", "Username": username, "Password": password, "Password_confirmation": password_confirmation})
        return
	}

    // DB への保存
    _, err = db.Exec("INSERT INTO users(name, password) VALUES (?, ?)", username, hash(password))
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }

    ctx.Redirect(http.StatusFound, "/login")
}

func LoginUserForm(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "Login"})
}

const userkey = "user"
 
func Login(ctx *gin.Context) {
    username := ctx.PostForm("username")
    password := ctx.PostForm("password")

    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }

    // ユーザの取得
    var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE name = ?", username)
    if err != nil {
        ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
        return
    }

    // パスワードの照合
    if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
        ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
        return
    }

    // セッションの保存
    session := sessions.Default(ctx)
    session.Set(userkey, user.ID)
    session.Save()

    ctx.Redirect(http.StatusFound, "/list")
}

func LoginCheck(ctx *gin.Context) {
    if sessions.Default(ctx).Get(userkey) == nil {
        ctx.Redirect(http.StatusFound, "/login")
        ctx.Abort()
    } else {
        ctx.Next()
    }
}

func Logout(ctx *gin.Context) {
    session := sessions.Default(ctx)
    session.Clear()
    session.Options(sessions.Options{MaxAge: -1})
    session.Save()
    ctx.Redirect(http.StatusFound, "/")
}

func EditUserForm(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "edit_user_form.html", gin.H{"Title": "Edit user"})
}

// ユーザー情報再登録のための関数
func ReregisterUser(ctx *gin.Context) {
    // フォームデータの受け取り
    current_username := ctx.PostForm("current-username")
    new_username := ctx.PostForm("new-username")
    current_password := ctx.PostForm("current-password")
    new_password := ctx.PostForm("new-password")
    new_password_confirmation := ctx.PostForm("new-password-confirmation")
    switch {
    case current_username == "":
        ctx.HTML(http.StatusBadRequest, "edit_user_form.html", gin.H{"Title": "Edit user", "Error": "Current usernane is not provided", "Current_username": current_username})
    case new_username == "":
        ctx.HTML(http.StatusBadRequest, "edit_user_form.html", gin.H{"Title": "Edit user", "Error": "New usernane is not provided", "New_username": new_username})
    case current_password == "":
        ctx.HTML(http.StatusBadRequest, "edit_user_form.html", gin.H{"Title": "Edit user", "Error": "Current password is not provided", "Current_password": current_password})
	case new_password == "":
		ctx.HTML(http.StatusBadRequest, "edit_user_form.html", gin.H{"Title": "Edit user", "Error": "New password is not provided", "New_password": new_password})
    case new_password_confirmation == "":
		ctx.HTML(http.StatusBadRequest, "edit_user_form.html", gin.H{"Title": "Edit user", "Error": "New password confirmation is not provided", "New_password_confirmation": new_password_confirmation})
    }
    
    // DB 接続
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }

    // 現在設定されているパスワードが正しいかどうかチェック
    var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE name = ?", current_username)
    if err != nil {
        ctx.HTML(http.StatusBadRequest, "edit_user_form.html", gin.H{"Title": "Edit user", "Current_username": current_username, "Error": "No such user"})
        return
    }
    if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(current_password)) {
        ctx.HTML(http.StatusBadRequest, "edit_user_form.html", gin.H{"Title": "Edit user", "Current_username": current_username, "Error": "Incorrect password"})
        return
    }


	// パスワードが適切かチェック
	if utf8.RuneCountInString(new_password) <= 5 {
		ctx.HTML(http.StatusBadRequest, "edit_user_form.html", gin.H{"Title": "Edit user", "Error": "Password must be at least 6 characters.", "Current_username": current_username, "New_username": new_username, "Current_password": current_password, "New_password": new_password, "New_password_confirmation": new_password_confirmation})
		return
	} 
	_, err = strconv.Atoi(new_password)
	if err == nil {
		ctx.HTML(http.StatusBadRequest, "edit_user_form.html", gin.H{"Title": "Edit user", "Error": "Do not use only numbers for password.", "Current_username": current_username, "New_username": new_username, "Current_password": current_password, "New_password": new_password, "New_password_confirmation": new_password_confirmation})
		return
	}
	
 
    // 重複チェック
    var duplicate int
    err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", new_username)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    if duplicate > 0 {
        ctx.HTML(http.StatusBadRequest, "edit_user_form.html", gin.H{"Title": "Register user", "Error": "New username is already taken", "Current_username": current_username, "New_username": new_username, "Current_password": current_password, "New_password": new_password, "New_password_confirmation": new_password_confirmation})
        return
    }

	// new_passwordとnew_password_confirmationが同じかチェック
	if new_password != new_password_confirmation {
		ctx.HTML(http.StatusBadRequest, "edit_user_form.html", gin.H{"Title": "Register user", "Error": "New password and New password confirmarion is not equal", "Current_username": current_username, "New_username": new_username, "Current_password": current_password, "New_password": new_password, "New_password_confirmation": new_password_confirmation})
        return
	}

    // DB の更新
	_, err = db.Exec("UPDATE users SET name=?, password=? WHERE id=?", new_username, hash(new_password), user.ID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

    ctx.Redirect(http.StatusFound, "/list")
}

// ユーザーアカウントの削除を実行
func DeleteUser(ctx *gin.Context) {
    userID := sessions.Default(ctx).Get(userkey)

    // DB 接続
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }

    // userIDをもとにusersDBからデータ削除
    _, err = db.Exec("DELETE FROM users WHERE id=?", userID)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }

    // userIDをもとにtasksDBからデータ削除
    _, err = db.Exec("DELETE FROM tasks WHERE id IN (SELECT task_id FROM ownership WHERE user_id=?)", userID)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }

    // userIDをもとにownershipDBからデータ削除
    _, err = db.Exec("DELETE FROM ownership WHERE user_id=?", userID)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    
    // ログアウト処理とトップページへのリダイレクト
    Logout(ctx)
} 