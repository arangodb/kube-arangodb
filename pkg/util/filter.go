package util

func NewFilter[T any](in []T) Filter[T] {
	return filterList[T](in)
}

type Filter[T any] interface {
	Filter(predicate func(in T) bool) Filter[T]
	Get() []T
}

type filterList[T any] []T

func (f filterList[T]) Filter(predicate func(in T) bool) Filter[T] {
	n := make(filterList[T], 0, len(f))

	for _, el := range f {
		if predicate(el) {
			n = append(n, el)
		}
	}

	return n
}

func (f filterList[T]) Get() []T {
	return f
}
