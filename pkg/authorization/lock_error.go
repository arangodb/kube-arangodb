package authorization

type AlreadyLocked struct {
}

func (a AlreadyLocked) Error() string {
	return "AlreadyLocked"
}
