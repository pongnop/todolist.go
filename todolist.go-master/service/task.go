package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	database "todolist.go/db"
)

// TaskList renders list of tasks in DB
func TaskList(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user") 
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Get query parameter
    kw := ctx.Query("kw")
	is_done := ctx.Query("is_done")

	//実行するSQLコマンドを作成
	query := fmt.Sprintf("SELECT id, title, created_at, is_done FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ? AND title LIKE '%s'", "%" + kw + "%")
	if is_done == "t"{
		query = query + " AND is_done = 1"
	} else if is_done == "f"{
		query = query + " AND is_done = 0"
	}

	// Get tasks in DB
	var tasks []database.Task
    err = db.Select(&tasks, query, userID)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }

	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks, "Kw": kw, "Is_done" : is_done})
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// parse ID given as a parameter
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get a task with given ID
	var task database.Task
	err = db.Get(&task, "SELECT id, title, created_at, is_done, user_id FROM tasks INNER JOIN ownership ON task_id = id WHERE id=?", id) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
    if task.UserID != userID{
        Error(http.StatusBadRequest, "You are not authorized for this task")(ctx)
        return
    }


	// Render task
	//ctx.String(http.StatusOK, task.Title)  // Modify it!!
	ctx.HTML(http.StatusOK, "task.html", task)
}

func NewTaskForm(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration"})
}

func RegisterTask(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
    // Get task title
    title, exist := ctx.GetPostForm("title")
    if !exist {
        Error(http.StatusBadRequest, "No title is given")(ctx)
        return
    }
    // Get DB connection
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
	// Create new data with given title on DB
	tx := db.MustBegin()
    result, err := db.Exec("INSERT INTO tasks (title) VALUES (?)", title)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
	taskID, err := result.LastInsertId()
    if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    _, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", userID, taskID)
    if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    tx.Commit()
    // Render status
	path := "/list"  // デフォルトではタスク一覧ページへ戻る
    if id, err := result.LastInsertId(); err == nil {
        path = fmt.Sprintf("/task/%d", id)   // 正常にIDを取得できた場合は /task/<id> へ戻る
    }
    ctx.Redirect(http.StatusFound, path)
}

func EditTaskForm(ctx *gin.Context) {
    userID := sessions.Default(ctx).Get("user")
    // ID の取得
    id, err := strconv.Atoi(ctx.Param("id"))
    if err != nil {
        Error(http.StatusBadRequest, err.Error())(ctx)
        return
    }
    // Get DB connection
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }

    // Get target task
    var task database.Task
    err = db.Get(&task, "SELECT id, title, created_at, is_done, user_id FROM tasks INNER JOIN ownership ON task_id = id WHERE id=?", id)
    if err != nil {
        Error(http.StatusBadRequest, err.Error())(ctx)
        return
    }
    if task.UserID != userID{
        Error(http.StatusBadRequest, "You are not authorized for this task")(ctx)
        return
    }
    // Render edit form
    ctx.HTML(http.StatusOK, "form_edit_task.html",
        gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task})
}

func UpdateTask(ctx *gin.Context){
	//parse ID given as a parameter
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	//Get task title
	title, exist := ctx.GetPostForm("title")
    if !exist {
        Error(http.StatusBadRequest, "No title is given")(ctx)
        return
    }
	//Get task status
	status, exist:= ctx.GetPostForm("is_done")
	if !exist {
        Error(http.StatusBadRequest, "No title is given")(ctx)
        return
    }
	is_done, _ := strconv.ParseBool(status)
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//Update task in DB
	_, err = db.Exec("UPDATE tasks SET title=?, is_done=? WHERE id=?", title, is_done,id)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
	// Render status
	path := fmt.Sprintf("/task/%d", id)
    ctx.Redirect(http.StatusFound, path)
}

func DeleteTask(ctx *gin.Context) {
    userID := sessions.Default(ctx).Get("user")
    // ID の取得
    id, err := strconv.Atoi(ctx.Param("id"))
    if err != nil {
        Error(http.StatusBadRequest, err.Error())(ctx)
        return
    }
    // Get DB connection
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    var task database.Task
    err = db.Get(&task, "SELECT user_id FROM tasks INNER JOIN ownership ON task_id = id WHERE id=?", id)
    if err != nil {
        Error(http.StatusBadRequest, err.Error())(ctx)
        return
    }
    if task.UserID != userID{
        Error(http.StatusBadRequest, "You are not authorized for this task")(ctx)
        return
    }
    // Delete the task from DB
    _, err = db.Exec("DELETE FROM tasks WHERE id=?", id)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    // Redirect to /list
    ctx.Redirect(http.StatusFound, "/list")
}