package myservice

import "github.com/go-sql-driver/mysql"

type serviceError struct {
	Code        uint16 `json:"code"`
	Error       string `json:"error"`
	Description string `json:"description"`
}

func newServiceError(code uint16, err string, description string) func(editErr ...string) *serviceError {
	return func(editError ...string) *serviceError {
		if len(editError) != 0 {
			err = editError[0]
		}
		return &serviceError{Code: code, Error: err, Description: description}
	}
}

func isDatabaseError(err error) (dberr *mysql.MySQLError, ok bool) {
	dberr, ok = (err).(*mysql.MySQLError)
	return dberr, ok
}

// 0-99 - неизвестные ошибки, данные ошибки летят в лог
var (
	errUnknownDatabase = newServiceError(1, "unknown database error", "неизвестная ошибка при работе с базой данных")
	errUnknownServer   = newServiceError(2, "unknown server error", "неизвестная ошибка сервера")
)

// 100-199 - ошибки связанные с некорректно введенными данными
var (
	errIncorrectEmail     = newServiceError(101, "incorrect email", "некорректный email адрес")
	errIncorrectQuery     = newServiceError(102, "incorrect query parametr", "некорректное значение параметра в запросе")
	errIncorrectInputData = newServiceError(103, "incorrect input data", "некорректно переданные данные")
)

// 200-299 - ошибки связанные с базой данных
var (
	errEmailExists   = newServiceError(201, "email already exists", "данный email уже сущетсвует")
	errEmailNotFound = newServiceError(202, "email not found", "данный email не найден")
)
