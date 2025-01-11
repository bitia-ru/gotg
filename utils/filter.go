package utils

func EmptyFilter(s any) bool {
	switch s := s.(type) {
	case string:
		return s == ""
	default:
		return s == nil
	}
}

func NotEmptyFilter(s any) bool {
	return !EmptyFilter(s)
}

func Filter[T any](ss []T, test func(any) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
