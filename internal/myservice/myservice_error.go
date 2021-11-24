package myservice

import "github.com/go-sql-driver/mysql"

type myServiceError struct {
	Code        uint16 `json:"code"`
	Error       string `json:"error"`
	Description string `json:"description"`
}

func newMyServiceError(code uint16, err string, description string) func(editErr ...string) *myServiceError {
	return func(editErr ...string) *myServiceError {
		if len(editErr) != 0 {
			err = editErr[0]
		}
		return &myServiceError{Code: code, Error: err, Description: description}
	}
}

// func newErr(code uint16, description string) func(string) *myServiceError {
// 	e := &myServiceError{Code: code, Description: description}
// 	return func(err string) *myServiceError {
// 		e.Error = err
// 		return e
// 	}
// }

func isDatabaseError(err error) (dberr *mysql.MySQLError, ok bool) {
	dberr, ok = (err).(*mysql.MySQLError)
	return dberr, ok
}

// 0-99 - неизвестные ошибки, данные ошибки летят в лог
var (
	ERR_UNKNOWN_DATABASE = newMyServiceError(1, "unknown database error", "неизвестная ошибка при работе с базой данных")
	ERR_UNKNOWN_SERVER   = newMyServiceError(2, "unknown server error", "неизвестная ошибка сервера")
)

// 100-199 - ошибки связанные с некорректно введенными данными
var (
	ERR_INCORRECT_EMAIL      = newMyServiceError(101, "incorrect email", "некорректный email адрес")
	ERR_INCORRECT_QUERY_TYPE = newMyServiceError(102, "incorrect query parametr `type`", "некорректное значение параметра `type` в запросе")
	ERR_INCORRECT_INPUT_DATA = newMyServiceError(103, "incorrect input data", "некорректно переданные данные")
)

// 200-299 - ошибки связанные с базой данных
var (
	ERR_EMAIL_ALREADY_EXISTS = newMyServiceError(201, "email already exists", "данный email уже сущетсвует")
	ERR_EMAIL_NOT_FOUND      = newMyServiceError(202, "email not found", "данный email не найден")
)
