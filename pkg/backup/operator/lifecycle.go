package operator

// LifecyclePreStart interface executed before operator starts
type LifecyclePreStart interface {
	Handler

	LifecyclePreStart() error
}

// ExecLifecyclePreStart execute PreStart step on handler if interface is implemented
func ExecLifecyclePreStart(handler Handler) error {
	if l, ok := handler.(LifecyclePreStart); ok {
		return l.LifecyclePreStart()
	}
	return nil
}
