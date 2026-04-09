package optional

// Optional is a container object which may or may not contain a non-zero value.
type Optional[T any] struct {
	value   T
	present bool
}

func Some[T any](value T) Optional[T] {
	return Optional[T]{
		value:   value,
		present: true,
	}
}

func None[T any]() Optional[T] {
	return Optional[T]{}
}

func (o Optional[T]) IsPresent() bool {
	return o.present
}

func (o Optional[T]) OrElse(other T) T {
	if o.present {
		return o.value
	}

	return other
}
