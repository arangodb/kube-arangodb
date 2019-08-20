package operator

type LifecyclePreStart interface {
	Handler

	LifecyclePreStart() error
}

func ExecLifecyclePreStart(handler Handler) error {
	if l, ok := handler.(LifecyclePreStart); ok {
		return l.LifecyclePreStart()
	}
	return nil
}
