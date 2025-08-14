package optional

import (
	"encoding/json"
	"testing"
)

func TestBasicSetGet(t *testing.T) {
	var f Field[int]
	if f.Present() {
		t.Fatal("expected not present")
	}
	f.Set(10)
	if v, ok := f.Get(); !ok || v != 10 {
		t.Fatalf("get mismatch: %v %v", v, ok)
	}
	if f.OrElse(5) != 10 {
		t.Fatal("orElse should return present value")
	}
	f.Clear()
	if f.Present() {
		t.Fatal("expected cleared to be absent")
	}
	if f.OrElseGet(func() int { return 7 }) != 7 {
		t.Fatal("orElseGet should use supplier when absent")
	}
}

func TestNoneAndNew(t *testing.T) {
	n := None[string]()
	if n.Present() {
		t.Fatal("none should be absent")
	}
	s := NewField("x")
	if !s.Present() || s.OrElse("") != "x" {
		t.Fatal("new field should be present with value")
	}
}

func TestAdoptPtrAndRef(t *testing.T) {
	x := 3
	f := AdoptPtr(&x)
	if !f.Present() {
		t.Fatal("expected present")
	}
	p, ok := f.Ref()
	if !ok || p != &x {
		t.Fatal("ref should return original pointer")
	}
	*p = 9
	if v, _ := f.Get(); v != 9 {
		t.Fatalf("adopt/ref should mutate underlying: %v", v)
	}
	// ToPtr should be a copy
	cp := f.ToPtr()
	if cp == &x {
		t.Fatal("ToPtr must not return internal pointer")
	}
}

func TestJSONMarshalUnmarshalAndOmitempty(t *testing.T) {
	type W struct {
		Name string       `json:"name"`
		Age  *Field[int]  `json:"age,omitempty"`
		Note *Field[*int] `json:"note,omitempty"`
	}
	var w W
	w.Name = "Ava"
	// Absent fields should be omitted (nil pointers)
	b, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"name":"Ava"}` {
		t.Fatalf("unexpected json: %s", b)
	}

	// Present zero value should be emitted
	w.Age = &Field[int]{}
	w.Age.Set(0)
	b, err = json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"name":"Ava","age":0}` {
		t.Fatalf("unexpected json: %s", b)
	}

	// Unmarshal whitespace null should clear (nil pointer)
	data := []byte("{\n  \"age\": null\n}")
	if err := json.Unmarshal(data, &w); err != nil {
		t.Fatal(err)
	}
	if w.Age != nil {
		t.Fatal("age pointer should be nil after null")
	}

	// Pointer typed T: present nil vs absent
	var nn *int
	f := NewField(nn)
	w.Note = &f
	b, err = json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"name":"Ava","note":null}` {
		t.Fatalf("unexpected json for nil pointer present: %s", b)
	}
}

func TestMapFlatMapFilter(t *testing.T) {
	f := NewField(10)
	g := Map(f, func(v int) int { return v + 1 })
	if v, _ := g.Get(); v != 11 {
		t.Fatalf("map result: %v", v)
	}
	h := FlatMap(g, func(v int) Field[int] { return NewField(v * 2) })
	if v, _ := h.Get(); v != 22 {
		t.Fatalf("flatmap result: %v", v)
	}
	i := Filter(h, func(v int) bool { return v%2 == 0 })
	if _, ok := i.Get(); !ok {
		t.Fatal("expected filter to keep even value")
	}
	j := Filter(h, func(v int) bool { return v%2 == 1 })
	if j.Present() {
		t.Fatal("expected filter to drop odd predicate")
	}
}
