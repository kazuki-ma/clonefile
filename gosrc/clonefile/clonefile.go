package clonefile

type Error struct {
	Unsupported bool
	errorString string
}

func (c *Error) Error() string {
	return c.errorString
}
