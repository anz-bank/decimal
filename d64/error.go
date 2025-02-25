package d64

type Error string

func (e Error) Error() string {
	return string(e)
}
