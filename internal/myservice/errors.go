package myservice

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/iivkis/pos.7-era.backend/internal/config"
)

var errlog *log.Logger

func init() {
	f, err := os.Create("err_unknown.log")
	if err != nil {
		panic(err)
	}
	errlog = log.New(f, fmt.Sprintf("[port: %s] ", *config.Flags.Port), 0)
}

type serviceError struct {
	Code  uint16 `json:"code"`
	Error string `json:"error"`
}

func newServiceError(code uint16, err string) func(...string) *serviceError {
	return func(description ...string) *serviceError {
		e := err
		if len(description) != 0 {
			e = err + ": " + description[0]
		}
		return &serviceError{Code: code, Error: e}
	}
}

func newServiceErrorLog(code uint16, err string) func(...string) *serviceError {
	return func(description ...string) *serviceError {
		log.Print(description)
		errlog.Print(time.Now().String(), description)
		return &serviceError{Code: code, Error: err}
	}
}

func isDatabaseError(err error) (dberr *mysql.MySQLError, ok bool) {
	dberr, ok = (err).(*mysql.MySQLError)
	return dberr, ok
}

// 0-99 - неизвестные ошибки, данные ошибки летят в лог
var (
	errUnknown = newServiceErrorLog(1, "unknown server error")
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
	errPermissionDenided = newServiceError(303, "permission denided")
)
