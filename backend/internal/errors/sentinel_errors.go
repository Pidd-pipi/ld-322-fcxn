package apperrors

import (
	"errors"
	"net/http"
)

var ErrRecordNotFound = errors.New("record not found")

// ErrAlertConflict 表示报警当前状态不允许人工确认：重复确认或已自动恢复。
var ErrAlertConflict = New(40901, "报警已确认或已恢复，无需重复处理", http.StatusConflict)
