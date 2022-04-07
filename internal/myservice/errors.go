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
//TODO: Сделать полет в лог и сам лог
var (
	errUnknownDatabase = newServiceError(1, "database error")
	errUnknownServer   = newServiceError(2, "server error")
)

// 100-199 - ошибки связанные с некорректно переданными данными
var (
	errIncorrectEmail       = newServiceError(101, "incorrect email")
	errIncorrectInputData   = newServiceError(103, "incorrect input data")
	errIncorrectConfirmCode = newServiceError(104, "incorrect confirm code")
	errUploadFile           = newServiceError(105, "upload file error")
)

// 200-299 - ошибки связанные с базой данных
var (
	errEmailExists         = newServiceError(201, "email already exists")
	errEmailNotFound       = newServiceError(202, "email not found")
	errRecordNotFound      = newServiceError(203, "record not found")
	errIncorrectPassword   = newServiceError(204, "invalid password")
	errRecordAlreadyExists = newServiceError(206, "the record already exists")
	errForeignKey          = newServiceError(207, "foreign key error")
)

//300-399 - ошибки связанные с токеном и доступом
var (
	errParsingJWT        = newServiceError(300, "jwt token parsing error")
	errUndefinedJWT      = newServiceError(301, "jwt token undefined in header `Authorization`")
	errNoAccessRights    = newServiceError(302, "no access rights")
	errPermissionDenided = newServiceError(303, "permission denided")
)
