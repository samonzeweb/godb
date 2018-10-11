package dberror

import "strings"

// UniqueConstraint error is for handling for unique constrait errors
type UniqueConstraint struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Err     error  `json:"err"`
}

func (e UniqueConstraint) Error() string {
	return e.Message
}

// CheckConstraint error is for handling for check constrait errors
type CheckConstraint struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Err     error  `json:"err"`
}

func (e CheckConstraint) Error() string {
	return e.Message
}

// ForeignKeyConstraint error is for handling for check constrait errors
type ForeignKeyConstraint struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Err     error  `json:"err"`
}

func (e ForeignKeyConstraint) Error() string {
	return e.Message
}

// ExtractStr is used to extract error message
func ExtractStr(s, left, right string) string {
	return strings.Split(strings.Split(s, left)[1], right)[0]
}
