package myservice

import "github.com/go-sql-driver/mysql"

type serviceError struct {
	Code  uint16 `json:"code"`
	Error string `json:"error"`
}

func newServiceError(code uint16, err string) func(editErr ...string) *serviceError {
	return func(editError ...string) *serviceError {
		e := err
		if len(editError) != 0 {
			e = err + ": " + editError[0]
		}
		return &serviceError{Code: code, Error: e}
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
	errIncorrectEmail       = newServiceError(101, "incorrect email")
	errIncorrectInputData   = newServiceError(103, "incorrect input data")
	errIncorrectConfirmCode = newServiceError(104, "incorrect confirm code")
)

// 200-299 - ошибки связанные с базой данных
var (
	errEmailExists       = newServiceError(201, "email already exists")
	errEmailNotFound     = newServiceError(202, "email not found")
	errRecordNotFound    = newServiceError(203, "record not found")
	errIncorrectPassword = newServiceError(204, "invalid password")
)

//300-399 - ошибки для внешнего импорта
var (
	ErrParsingJWT   = newServiceError(300, "jwt token parsing error")
	ErrUndefinedJWT = newServiceError(301, "jwt token undefined in header `Authorization`")
)
