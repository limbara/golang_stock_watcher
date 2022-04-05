package customerrors

import "fmt"

var (
	ResponseErrorServer     = ResponseError{"Something Went Wrong!", 500, nil, nil}
	ResponseNotFound        = ResponseError{"Resource Not Found", 404, nil, nil}
	ResponseValidationError = ResponseError{"Validation Error", 400, nil, nil}
)

type ResponseError struct {
	Message  string      `json:"message"`
	HttpCode int         `json:"http_code"`
	Err      error       `json:"error"`
	Data     interface{} `json:"data"`
}

func (re *ResponseError) Error() string {
	return fmt.Sprintf("ResponseError( Message: %s, HttpCode: %d, Err: %s, Data: %+v", re.Message, re.HttpCode, re.Err, re.Data)
}

func (re *ResponseError) Unwrap() error {
	return re.Err
}

func NewResponseError(message string, httpCode int, err error) ResponseError {
	return ResponseError{
		Message:  message,
		HttpCode: httpCode,
		Err:      err,
	}
}

func (re *ResponseError) SetError(err error) *ResponseError {
	re.Err = err

	return re
}

func (re *ResponseError) SetMessage(message string) *ResponseError {
	re.Message = message

	return re
}

func (re *ResponseError) SetHttpCode(httpCode int) *ResponseError {
	re.HttpCode = httpCode

	return re
}

func (re *ResponseError) SetData(data interface{}) *ResponseError {
	re.Data = data

	return re
}
