package docker

import "testing"

func TestStatusSuccess(t *testing.T) {
	var crctstatus = []struct {
		a int
	}{
		{200},
		{239},
		{278},
		{300},
		{399},
	}
	var wrngstatus = []struct {
		b int
	}{
		{1},
		{50},
		{111},
		{199},
		{400},
		{1000},
	}

	for _, i := range crctstatus {
		res := statusSuccess(i.a)
		if res != true {
			t.Errorf("%d is an invalid status", i.a)
		}

		for _, j := range wrngstatus {
			res := statusSuccess(j.b)
			if res == true {
				t.Errorf("%d is an invalid status", j.b)
			}

		}
	}
}
