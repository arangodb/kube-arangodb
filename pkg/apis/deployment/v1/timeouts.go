package v1

import (
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	addMemberTimeout = time.Minute * 5
)

type Timeouts struct {
	AddMember *Timeout `json:"addMember,omitempty"`
}

func (t *Timeouts) Get() Timeouts {
	if t == nil {
		return Timeouts{}
	}

	return *t
}

type Timeout meta.Duration

func (t *Timeout) Get(d time.Duration) time.Duration {
	if t == nil {
		return d
	}

	return t.Duration
}
