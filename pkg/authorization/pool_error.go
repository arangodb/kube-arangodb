package authorization

type PoolOutOfBoundsError struct{}

func (PoolOutOfBoundsError) Error() string {
	return "pool out of bounds"
}
