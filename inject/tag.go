package inject

import "strings"

type tag struct {
	key      string
	optional bool
}

func parseTags(value string) tag {
	key, ok := strings.CutSuffix(value, ";optional")
	return tag{
		key:      key,
		optional: ok,
	}
}
