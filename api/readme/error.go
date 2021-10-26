package readme

type Error struct {
	ErrorCode string `json:"error"`
	Message   string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}
