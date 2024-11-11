package api

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	group "github.com/Drelf2018/gin-group"
	"github.com/Drelf2018/webhook/model"
	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidOp   = errors.New("webhook/api: invalid patch op")
	ErrInvalidPath = errors.New("webhook/api: invalid patch path")
	ErrPermDenied  = errors.New("webhook/api: permission denied")
	ErrMultipleErr = errors.New("webhook/api: multiple errors")
)

type PatchBody struct {
	Op    string `json:"op"` // [replace, add, remove, move, copy, test]
	Path  string `json:"path"`
	Value string `json:"value,omitempty"`
	From  string `json:"from,omitempty"`
}

func PatchUser(ctx *gin.Context) (any, error) {
	var body []PatchBody
	err := ctx.ShouldBindJSON(&body)
	if err != nil {
		return 1, err
	}

	user := &model.User{UID: ctx.Param("uid")}
	tx := UserDB().Limit(1).Find(user)
	if tx.Error != nil {
		return 2, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 3, ErrUserNotExist
	}

	me := &model.User{UID: GetUID(ctx)}
	err = UserDB().First(me).Error
	if err != nil {
		return 4, err
	}

	var errs []group.Response
	for i, patch := range body {
		err = nil
		switch patch.Path {
		default:
			err = ErrInvalidPath
		case "/role":
			err = PatchUserRole(ctx, me, user, patch)
		case "/ban":
			err = PatchUserBan(ctx, me, user, patch)
		case "/name":
			err = PatchUserName(ctx, me, user, patch)
		case "/nickname":
			err = PatchUserNickname(ctx, me, user, patch)
		}
		if err != nil {
			errs = append(errs, group.Response{Code: i, Error: err.Error()})
		}
	}
	if len(errs) != 0 {
		return errs, group.E(5, ErrMultipleErr)
	}

	err = UserDB().Save(user).Error
	if err != nil {
		return 6, err
	}
	return "success", nil
}

func PatchUserRole(ctx *gin.Context, me, user *model.User, patch PatchBody) error {
	if !me.Role.IsAdmin() {
		return ErrPermDenied
	}
	if me.Role <= user.Role {
		return ErrPermDenied
	}
	switch patch.Op {
	case "replace":
		i, err := strconv.Atoi(patch.Value)
		if err != nil {
			return err
		}
		if i <= 0 || i >= int(me.Role) {
			return ErrPermDenied
		}
		user.Role = model.Role(i)
		return nil
	default:
		return ErrInvalidOp
	}
}

func PatchUserBan(ctx *gin.Context, me, user *model.User, patch PatchBody) error {
	if !me.Role.IsAdmin() {
		return ErrPermDenied
	}
	if me.Role <= user.Role {
		return ErrPermDenied
	}
	defer func() {
		if user.Ban.After(time.Now()) {
			tokenIssuedAt.Delete(user.UID)
		}
	}()
	switch patch.Op {
	case "replace":
		return user.Ban.UnmarshalJSON([]byte(patch.Value))
	case "add":
		i, err := strconv.Atoi(patch.Value)
		if err != nil {
			return err
		}
		user.Ban = user.Ban.Add(time.Duration(i))
		return nil
	case "remove":
		user.Ban = time.Time{}
		return nil
	default:
		return ErrInvalidOp
	}
}

func PatchUserName(ctx *gin.Context, me, user *model.User, patch PatchBody) error {
	if !me.Role.IsAdmin() {
		return ErrPermDenied
	}
	if me.Role < user.Role {
		return ErrPermDenied
	}
	if me.Role == user.Role && me.UID != user.UID {
		return ErrPermDenied
	}
	switch patch.Op {
	case "replace":
		user.Name = patch.Value
		return nil
	default:
		return ErrInvalidOp
	}
}

func PatchUserNickname(ctx *gin.Context, me, user *model.User, patch PatchBody) error {
	if me.UID != user.UID && (!me.Role.IsAdmin() || me.Role <= user.Role) {
		return ErrPermDenied
	}
	switch patch.Op {
	case "replace":
		user.Nickname = patch.Value
		return nil
	default:
		return ErrInvalidOp
	}
}

func PatchTaskID(ctx *gin.Context) (any, error) {
	var body []PatchBody
	err := ctx.ShouldBindJSON(&body)
	if err != nil {
		return 1, err
	}

	task := &model.Task{}
	tx := UserDB().Limit(1).Find(task, "id = ? AND user_id = ?", ctx.Param("id"), GetUID(ctx))
	if tx.Error != nil {
		return 2, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 3, ErrTaskNotExist
	}
	err = task.Api.Header.Unwrap()
	if err != nil {
		return 4, err
	}

	var errs []group.Response
	for i, patch := range body {
		err = nil
		switch patch.Path {
		default:
			err = ErrInvalidPath
		case "/name":
			task.Name = patch.Value
		case "/method":
			task.Api.Method = patch.Value
		case "/url":
			task.Api.URL = patch.Value
		case "/body":
			task.Api.Body = patch.Value
		case "/enable":
			task.Enable, err = strconv.ParseBool(patch.Value)
		case "/request_once":
			task.RequestOnce, err = strconv.ParseBool(patch.Value)
		case "/header":
			err = json.Unmarshal([]byte(patch.Value), &task.Api.Header)
		case "/parameter":
			err = json.Unmarshal([]byte(patch.Value), &task.Api.Parameter)
		case "/submitter":
			err = json.Unmarshal([]byte(patch.Value), &task.Filter.Submitter)
		case "/platform":
			err = json.Unmarshal([]byte(patch.Value), &task.Filter.Platform)
		case "/type":
			err = json.Unmarshal([]byte(patch.Value), &task.Filter.Type)
		case "/uid":
			err = json.Unmarshal([]byte(patch.Value), &task.Filter.UID)
		}
		if err != nil {
			errs = append(errs, group.Response{Code: i, Error: err.Error()})
		}
	}
	if len(errs) != 0 {
		return errs, group.E(4, ErrMultipleErr)
	}

	err = UserDB().Save(task).Error
	if err != nil {
		return 5, err
	}
	return "success", nil
}
