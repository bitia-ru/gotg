package utils

func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func PanicOnErrorWrap[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}
