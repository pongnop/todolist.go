package service
 
import (
	"crypto/sha256"
	"encoding/hex"
    "net/http"
 
    "github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	database "todolist.go/db"
)
func hash(pw string) []byte {
    const salt = "todolist.go#"
    h := sha256.New()
    h.Write([]byte(salt))
    h.Write([]byte(pw))
    return h.Sum(nil)
}
 
func NewUserForm(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "new_user_form.html", gin.H{"Title": "Register user"})
}

func RegisterUser(ctx *gin.Context) {
    // フォームデータの受け取り
    username := ctx.PostForm("username")
    password := ctx.PostForm("password")
	password_confilm := ctx.PostForm("password_confilm")
    switch {
    case username == "":
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Usernane is not provided", "Username": username})
    case password == "":
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Password": password})
    case password_confilm == "":
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Confirm Password is not provided", "Password_confilm": password_confilm})
	}

	//確認用パスワードの確認
	if password != password_confilm{
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "確認用パスワードが正しくありません。", "Username": username, "Password": password})
        return
	}

    // DB 接続
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
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
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username, "Password": password, "Password_confilm": password_confilm})
        return
    }
    // DB への保存
    result, err := db.Exec("INSERT INTO users(name, password) VALUES (?, ?)", username, hash(password))
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
 
    // 保存状態の確認
    id, _ := result.LastInsertId()
    var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", id)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    //ctx.JSON(http.StatusOK, user)

    session := sessions.Default(ctx)
    session.Set(userkey, user.ID)
    session.Save()
 
    ctx.Redirect(http.StatusFound, "/")
    
    
}
func LoginForm(ctx *gin.Context){
	ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "Register user"})
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
    err = db.Get(&user, "SELECT id, name, password, is_deleted FROM users WHERE name = ?", username)
    if err != nil {
        ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
        return
    }
 
    // パスワードの照合
    if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
        ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
        return
    }
	//アカウントの利用可能状況を確認
	if user.Is_deleted {
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "This account has been deleted"})
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

func DeleteUser(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//Update DB
	_, err = db.Exec("UPDATE users SET is_deleted=1 WHERE id=?", userID)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    ctx.Redirect(http.StatusFound, "/logout")
}

func DeleteUserForm(ctx *gin.Context) {
    userID := sessions.Default(ctx).Get("user")
    // Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
    var user database.User
    _ = db.Get(&user, "SELECT name FROM users WHERE id = ?", userID)
    
    ctx.HTML(http.StatusOK, "delete_user_form.html", gin.H{"Title": "Delete user", "Username":user.Name})
}
func UserEditForm(ctx *gin.Context){
    userID := sessions.Default(ctx).Get("user")
    // Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
    var user database.User
    _ = db.Get(&user, "SELECT name FROM users WHERE id = ?", userID)
    ctx.HTML(http.StatusOK, "form_edit_user.html", gin.H{"Title": "ユーザー情報の編集", "New_username": user.Name})
}

func UpdateUser(ctx *gin.Context) {
    // フォームデータの受け取り
    new_username := ctx.PostForm("new_username")
    original_password := ctx.PostForm("original_password")
    new_password := ctx.PostForm("new_password")
	new_password_confilm := ctx.PostForm("new_password_confilm")
    //元のユーザー情報の取得
    userID := sessions.Default(ctx).Get("user")
        // Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
    var user database.User
    _ = db.Get(&user, "SELECT name, password FROM users WHERE id = ?", userID)

	//確認用パスワードの確認
	if new_password != new_password_confilm{
		ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "ユーザー情報の編集", "Error": "確認用パスワードが正しくありません。", "New_username": new_username, "Original_password": original_password, "New_password": new_password})
        return
	}

	// 重複チェック
    var duplicate int
    err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", new_username)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    if duplicate > 0 && new_username != user.Name{
        ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "ユーザー情報の編集", "Error": "Username is already taken", "New_username": new_username, "Original_password": original_password, "New_password": new_password, "New_password_confilm": new_password_confilm})
        return
    }
    //元のパスワードの照合
    if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(original_password)){
		ctx.HTML(http.StatusBadRequest, "form_edit_user.html", gin.H{"Title": "ユーザー情報の編集", "Error": "元のパスワードが正しくありません。", "New_username": new_username,"New_password": new_password, "New_password_confilm": new_password_confilm})
        return
	}
    // DB への保存
    _, err = db.Exec("UPDATE users SET name = ?, password = ? WHERE id = ?", new_username, hash(new_password), userID)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    
 
    ctx.Redirect(http.StatusFound, "/")
    
    
}