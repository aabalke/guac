package utils

import "github.com/hajimehoshi/ebiten/v2"

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

func MakeKeyUnique(in *[]ebiten.Key) {
	set := make(map[ebiten.Key]struct{}, len(*in))
	for _, b := range *in {
		set[b] = struct{}{}
	}

	out := make([]ebiten.Key, 0, len(set))
	for b := range set {
		out = append(out, b)
	}

	*in = out
}

func AppendKeyUnique(dst, src []ebiten.Key) []ebiten.Key {
outer:
	for i := range src {
		for j := range dst {
			if src[i] == dst[j] {
				continue outer
			}
		}

		dst = append(dst, src[i])

	}

	return dst
}

func AppendButtonUnique(dst, src []ebiten.StandardGamepadButton) []ebiten.StandardGamepadButton {
outer:
	for i := range src {
		for j := range dst {
			if src[i] == dst[j] {
				continue outer
			}
		}

		dst = append(dst, src[i])

	}

	return dst
}
