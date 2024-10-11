package utils

func ReduceSlice[T any, U any](sT *[]T, f func(accumulator U, current T) U) *U {
	var a U
	for _, t := range *sT {
		a = f(a, t)
	}
	return &a
}

func FilterSlice[T any](sT *[]T, f func(T) bool) (result []T) {
	for _, t := range *sT {
		if f(t) {
			result = append(result, t)
		}
	}
	return
}

func Find[T any](sT *[]T, f func(T) bool) (result T) {
	for _, t := range *sT {
		if f(t) {
			return t
		}
	}
	return
}
