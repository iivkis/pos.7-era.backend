package myservice

import "github.com/go-sql-driver/mysql"

type serviceError struct {
	Code  uint16 `json:"code"`
	Error string `json:"error"`
}

func newServiceError(code uint16, err string) func(editErr ...string) *serviceError {
	return func(editError ...string) *serviceError {
		if len(editError) != 0 {
			err = err + ": " + editError[0]
		}
		return &serviceError{Code: code, Error: err}
	}
}

func isDatabaseError(err error) (dberr *mysql.MySQLError, ok bool) {
	dberr, ok = (err).(*mysql.MySQLError)
	return dberr, ok
}

// 0-99 - неизвестные ошибки, данные ошибки летят в лог
var (
	errUnknownDatabase = newServiceError(1, "unknown database error")
	errUnknownServer   = newServiceError(2, "unknown server error")
)

// 100-199 - ошибки связанные с некорректно переданными данными
var (
	errIncorrectEmail     = newServiceError(101, "incorrect email")
	errIncorrectInputData = newServiceError(103, "incorrect input data")

	errIncorrectConfirmCode = newServiceError(104, "incorrect confirm code")
)

// 200-299 - ошибки связанные с базой данных
var (
	errEmailExists   = newServiceError(201, "email already exists")
	errEmailNotFound = newServiceError(202, "email not found")
)
