package myservice

type myServiceError struct {
	Code        uint16 `json:"code"`
	Error       string `json:"error"`
	Description string `json:"description"`
}

func newErr(code uint16, description string) func(string) *myServiceError {
	e := &myServiceError{Code: code, Description: description}
	return func(err string) *myServiceError {
		e.Error = err
		return e
	}
}

// 0-99 - неизвестные ошибки, данные ошибки летят в лог
var ()

// 100-199 - ошибки связанные с некорректно введенными данными
var (
	errIncorrectEmail     = newErr(101, "некорректный email адрес")
	errIncorrectQueryType = newErr(102, "некорректное значение параметра `type` в запросе")
	errIncorrectInputData = newErr(103, "некорректно переданные данные")
)

// 200-299 - ошибки связанные с базой данных
var ()
