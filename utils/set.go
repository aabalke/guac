package utils

type Set[T comparable] map[T]struct{}

func (s Set[T]) add(t T) {
	s[t] = struct{}{}
}

func (s Set[T]) has(t T) bool {
	_, ok := s[t]
	return ok
}

func MakeUnique[T comparable](in *[]T) {

	set := make(map[T]struct{}, len(*in))
	for _, b := range *in {
		set[b] = struct{}{}
	}

	out := make([]T, 0, len(set))
	for b := range set {
		out = append(out, b)
	}

	*in = out
}
