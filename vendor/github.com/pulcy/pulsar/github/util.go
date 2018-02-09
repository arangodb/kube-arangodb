package github

import "time"

/* oracle nvl, return first non-empty string */
func nvls(xs ...string) string {
	for _, s := range xs {
		if s != "" {
			return s
		}
	}

	return ""
}

// formats time `t` as `fmt` if it is not nil, otherwise returns `def`
func timeFmtOr(t *time.Time, fmt, def string) string {
	if t == nil {
		return def
	}
	return t.Format(fmt)
}
