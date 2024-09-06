package lib

func Must1[T any](x T, err error) T {
	if err != nil {
		panic(err)
	}
	return x
}
