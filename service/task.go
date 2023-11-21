package service

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
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
    display_isdone := ctx.Query("display_isdone")
    tag := ctx.Query("tag")

    // Get tasks in DB
    var tasks []database.Task
    // task-listで表示したいカラムが変更されたときにここも変更する
    query := "SELECT id, title, created_at, is_done, deadline, priority, tag, overview FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ?"
    switch {
    case kw != "":
        if display_isdone == "done" {
            if tag == "all" {
                err = db.Select(&tasks, query + " AND title LIKE ? AND is_done=b'1'", userID, "%" + kw + "%")
            } else {
                err = db.Select(&tasks, query + " AND title LIKE ? AND is_done=b'1' AND tag=?", userID, "%" + kw + "%", tag)
            }
        } else if display_isdone == "notdone"{
            if tag == "all" {
                err = db.Select(&tasks, query + " AND title LIKE ? AND is_done=b'0'", userID, "%" + kw + "%")
            } else {
                err = db.Select(&tasks, query + " AND title LIKE ? AND is_done=b'0' AND tag=?", userID, "%" + kw + "%", tag)
            }
        } else{
            if tag == "all" {
                err = db.Select(&tasks, query + " AND title LIKE ?", userID, "%" + kw + "%")
            } else {
                err = db.Select(&tasks, query + " AND title LIKE ? AND tag=?", userID, "%" + kw + "%", tag)
            }
        }
    default:
        if display_isdone == "done" {
            if tag == "all" {
                err = db.Select(&tasks, query + " AND is_done=b'1'", userID)
            } else {
                err = db.Select(&tasks, query + " AND is_done=b'1' AND tag=?", userID, tag)
            }
        } else if display_isdone == "notdone"{
            if tag == "all" {
                err = db.Select(&tasks, query + " AND is_done=b'0'", userID)
            } else {
                err = db.Select(&tasks, query + " AND is_done=b'0' AND tag=?", userID, tag)
            }
        } else{
            // 一番初めのGETはtagは""になるので、その時は全部表示できるようにする
            if tag == "all" || tag == ""{
                err = db.Select(&tasks, query, userID)
            } else {
                err = db.Select(&tasks, query + " AND tag=?", userID, tag)
            }
        }
    }
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    now := time.Now()
    // Render tasks
    ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks, "Kw": kw, "Now": now})
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
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
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
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
	var result sql.Result
	var err error
    userID := sessions.Default(ctx).Get("user")
    // Get task title
    title, exist := ctx.GetPostForm("title")
    if !exist {
        Error(http.StatusBadRequest, "No title is given")(ctx)
        return
    }
    // Get task deadline
    deadline, exist := ctx.GetPostForm("deadline")
    if !exist {
        Error(http.StatusBadRequest, "No deadline is given")(ctx)
        return
    }
    // deadlineに一日分足す
    deadline += " 23:59:59"
    // Get task priority
    priority, exist := ctx.GetPostForm("priority")
    if !exist {
        Error(http.StatusBadRequest, "No priority is given")(ctx)
        return
    }
    // Get task tag
    tag, exist := ctx.GetPostForm("tag")
    if !exist {
        Error(http.StatusBadRequest, "No tag is given")(ctx)
        return
    }
    // Get task overview
	overview, _ := ctx.GetPostForm("overview")
    // Get DB connection
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }

    tx := db.MustBegin()
    // Create new data with given title on DB
	if overview == "" {
		result, err = tx.Exec("INSERT INTO tasks (title, deadline, priority, tag) VALUES (?, ?, ?, ?)", title, deadline, priority, tag)
		if err != nil {
            tx.Rollback()
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
	} else {
		result, err = tx.Exec("INSERT INTO tasks (title, deadline, priority, tag, overview) VALUES (?, ?, ?, ?, ?)", title, deadline, priority, tag, overview)
		if err != nil {
            tx.Rollback()
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}
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
    ctx.Redirect(http.StatusFound, fmt.Sprintf("/task/%d", taskID))
}

func EditTaskForm(ctx *gin.Context) {
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
    err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
    if err != nil {
        Error(http.StatusBadRequest, err.Error())(ctx)
        return
    }
    // Render edit form
    ctx.HTML(http.StatusOK, "form_edit_task.html",
        gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task})
}

func UpdateTask(ctx *gin.Context) {

	// Get task id
	id, err := strconv.Atoi(ctx.Param("id"))
    if err != nil {
        Error(http.StatusBadRequest, err.Error())(ctx)
        return
    }
    // Get task title
    title, exist := ctx.GetPostForm("title")
    if !exist {
        Error(http.StatusBadRequest, "No title is given")(ctx)
        return
    }
	// Get task is_done
	is_done_str, exist := ctx.GetPostForm("is_done")
	if !exist {
		Error(http.StatusBadRequest, "No is_done value is given")(ctx)
		return
	}
	is_done, err := strconv.ParseBool(is_done_str)
	if err != nil{
		Error(http.StatusBadRequest, err.Error())(ctx)
        return
	}
	// Get task deadline
    deadline, exist := ctx.GetPostForm("deadline")
    if !exist {
        Error(http.StatusBadRequest, "No deadline is given")(ctx)
        return
    }
    // deadlineに一日分足す
    deadline += " 23:59:59"
    // Get task priority
    priority, exist := ctx.GetPostForm("priority")
    if !exist {
        Error(http.StatusBadRequest, "No priority is given")(ctx)
        return
    }
    // Get task tag
    tag, exist := ctx.GetPostForm("tag")
    if !exist {
        Error(http.StatusBadRequest, "No tag is given")(ctx)
        return
    }
	// Get task overview
	overview, exist := ctx.GetPostForm("overview")
	if !exist {
        Error(http.StatusBadRequest, "No overview is given")(ctx)
        return
    }
    // Get DB connection
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    // Create new data with given title on DB
	_, err = db.Exec("UPDATE tasks SET title=?, is_done=?, deadline=?, priority=?, tag=?, overview=? WHERE id=?", title, is_done, deadline, priority, tag, overview, id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	
    // Render status
    path := fmt.Sprintf("/task/%d", id)   // 正常にIDを取得できた場合は /task/<id> へ戻る
    ctx.Redirect(http.StatusFound, path)
}

func DeleteTask(ctx *gin.Context) {
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
    // Delete the task from DB
    _, err = db.Exec("DELETE FROM tasks WHERE id=?", id)
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    // Redirect to /list
    ctx.Redirect(http.StatusFound, "/list")
}

func TaskCheck(ctx *gin.Context) {
    userID_session := sessions.Default(ctx).Get("user")
    // Get DB connection
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        ctx.Abort()
        return
    }
    // parse ID given as a parameter
	taskID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
        ctx.Abort()
		return
	}

    // ownershipからtaskIDをもとにuserIDをとりだす。
	var ownership database.Ownership
	err = db.Get(&ownership, "SELECT user_id FROM ownership WHERE task_id=?", taskID)
    if err != nil {
        Error(http.StatusBadRequest, err.Error())(ctx)
        ctx.Abort()
        return
    }
    
    if userID_session !=  ownership.UserID{
        ctx.Redirect(http.StatusFound, "/login")
        ctx.Abort()
    } else {
        ctx.Next()
    }
}
