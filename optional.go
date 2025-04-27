package optional

import (
	"encoding/json"
)

type Field[T any] struct {
	value *T
}

func NewField[T any](v T) Field[T] {
	return Field[T]{&v}
}

func NewFieldFromPtr[T any](v *T) Field[T] {
	if v == nil {
		return Field[T]{}
	}
	return NewField[T](*v)
}

func (i *Field[T]) Set(v T) {
	i.value = &v
}

func (i *Field[T]) ToPtr() *T {
	if !i.Present() {
		return nil
	}
	v := *i.value
	return &v
}

func (i *Field[T]) Get() (T, bool) {
	if !i.Present() {
		var zero T
		return zero, false
	}
	return *i.value, true
}

func (i *Field[T]) MustGet() T {
	if !i.Present() {
		panic("value not present")
	}
	return *i.value
}

func (i *Field[T]) Present() bool {
	return i.value != nil
}

func (i *Field[T]) OrElse(v T) T {
	if i.Present() {
		return *i.value
	}
	return v
}

func (i *Field[T]) If(fn func(T)) {
	if i.Present() {
		fn(*i.value)
	}
}

func (i *Field[T]) MarshalJSON() ([]byte, error) {
	if i.Present() {
		return json.Marshal(i.value)
	}
	return json.Marshal(nil)
}

func (i *Field[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		i.value = nil
		return nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	i.value = &value
	return nil
}
