package main

import (
	"math"
	"reflect"
	"testing"
)

type Simple struct {
	KeyInt     int
	KeyFloat   float64
	KeyBool    bool
	KeyComplex complex64
	KeyString  string
	// KeyUint    uint64
}

type IDBlock struct {
	ID int
}

type Complex struct {
	SubSimple  Simple
	ManySimple []Simple
	Blocks     []IDBlock
}

type ErrorCase struct {
	Result   interface{}
	JsonData string
}

func TestMapStructFieldsByName(t *testing.T) {
	t.Run("valid struct", func(t *testing.T) {
		s := Simple{}
		v := reflect.ValueOf(&s)
		fields, err := mapStructFieldsByName(v)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(fields) != 5 {
			t.Errorf("expected 5 fields, got %d", len(fields))
		}
	})

	t.Run("non-struct", func(t *testing.T) {
		var i int
		_, err := mapStructFieldsByName(reflect.ValueOf(&i))
		if err == nil {
			t.Error("expected error for non-struct type")
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var s *Simple
		_, err := mapStructFieldsByName(reflect.ValueOf(s))
		if err == nil {
			t.Error("expected error for nil pointer")
		}
	})
}

func TestAssignSimpleValue(t *testing.T) {
	tests := []struct {
		name     string
		dst      interface{}
		src      interface{}
		expected interface{}
		wantErr  bool
	}{
		// Int tests
		{"int to int", 0, 42, 42, false},
		{"uint to int", 0, uint(42), 42, false},
		{"float to int", 0, 42.0, 42, false},
		{"large uint to int", 0, uint(math.MaxUint64), nil, true},
		{"string to int", 0, "42", nil, true},

		// Uint tests
		{"int to uint", uint(0), 42, uint(42), false},
		{"uint to uint", uint(0), uint(42), uint(42), false},
		{"float to uint", uint(0), 42.0, uint(42), false},
		{"negative to uint", uint(0), -1, nil, true},

		// Float tests
		{"int to float", 0.0, 42, 42.0, false},
		{"uint to float", 0.0, uint(42), 42.0, false},
		{"float to float", 0.0, 42.5, 42.5, false},

		// Complex tests
		{"complex to complex", complex64(0), complex128(1 + 2i), complex64(1 + 2i), false},
		{"int to complex", complex64(0), 42, nil, true},

		// Bool tests
		{"bool to bool", false, true, true, false},
		{"int to bool", false, 1, nil, true},

		// String tests
		{"string to string", "", "test", "test", false},
		{"int to string", "", 42, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := reflect.New(reflect.TypeOf(tt.dst)).Elem()
			src := reflect.ValueOf(tt.src)

			err := assignSimpleValue(dst, src)
			if (err != nil) != tt.wantErr {
				t.Errorf("assignSimpleValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(dst.Interface(), tt.expected) {
				t.Errorf("assignSimpleValue() = %v, want %v", dst.Interface(), tt.expected)
			}
		})
	}
}

func TestAssignArraySliceValue(t *testing.T) {
	t.Run("valid slice conversion", func(t *testing.T) {
		src := []interface{}{1, 2, 3}
		var dst []int

		err := assignArraySliceValue(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dst) != 3 || dst[0] != 1 || dst[1] != 2 || dst[2] != 3 {
			t.Errorf("unexpected result: %v", dst)
		}
	})

	t.Run("invalid src type", func(t *testing.T) {
		var dst []int
		err := assignArraySliceValue(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(42))
		if err == nil {
			t.Error("expected error for non-slice src")
		}
	})

	t.Run("invalid dst type", func(t *testing.T) {
		src := []interface{}{1, 2, 3}
		var dst int
		err := assignArraySliceValue(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src))
		if err == nil {
			t.Error("expected error for non-slice dst")
		}
	})

	t.Run("nested slices", func(t *testing.T) {
		src := [][]interface{}{{"a", "b"}, {"c"}}
		var dst [][]string

		err := assignArraySliceValue(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dst) != 2 || len(dst[0]) != 2 || len(dst[1]) != 1 {
			t.Errorf("unexpected result: %v", dst)
		}
	})
}

func TestAssignMap(t *testing.T) {
	t.Run("valid map to struct", func(t *testing.T) {
		src := map[string]interface{}{
			"KeyInt":     42,
			"KeyString":  "test",
			"KeyBool":    true,
			"KeyFloat":   3.14,
			"KeyComplex": complex(1, 2),
			"KeyUint":    uint64(100),
		}
		var dst Simple

		err := assignMap(reflect.ValueOf(src), reflect.ValueOf(&dst).Elem())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dst.KeyInt != 42 || dst.KeyString != "test" {
			t.Errorf("unexpected result: %+v", dst)
		}
	})

	t.Run("invalid map key type", func(t *testing.T) {
		src := map[int]interface{}{1: "test"}
		var dst Simple
		err := assignMap(reflect.ValueOf(src), reflect.ValueOf(&dst).Elem())
		if err == nil {
			t.Error("expected error for non-string map key")
		}
	})

	t.Run("nil map", func(t *testing.T) {
		var src map[string]interface{}
		var dst Simple
		err := assignMap(reflect.ValueOf(src), reflect.ValueOf(&dst).Elem())
		if err != nil {
			t.Errorf("unexpected error for nil map: %v", err)
		}
	})
}

func TestI2S(t *testing.T) {
	t.Run("simple struct", func(t *testing.T) {
		src := map[string]interface{}{
			"KeyInt":     42,
			"KeyString":  "test",
			"KeyBool":    true,
			"KeyFloat":   3.14,
			"KeyComplex": complex(1, 2),
			"KeyUint":    uint64(100),
		}
		var dst Simple

		err := i2s(src, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dst.KeyInt != 42 || dst.KeyString != "test" {
			t.Errorf("unexpected result: %+v", dst)
		}
	})

	t.Run("complex struct", func(t *testing.T) {
		src := map[string]interface{}{
			"SubSimple": map[string]interface{}{
				"KeyInt":     10,
				"KeyString":  "sub",
				"KeyBool":    false,
				"KeyFloat":   1.1,
				"KeyComplex": complex(0, 1),
				"KeyUint":    uint64(50),
			},
			"ManySimple": []interface{}{
				map[string]interface{}{
					"KeyInt":     20,
					"KeyString":  "elem1",
					"KeyBool":    true,
					"KeyFloat":   2.2,
					"KeyComplex": complex(1, 0),
					"KeyUint":    uint64(60),
				},
			},
			"Blocks": []interface{}{
				map[string]interface{}{"ID": 1},
				map[string]interface{}{"ID": 2},
			},
		}
		var dst Complex

		err := i2s(src, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dst.ManySimple) != 1 || len(dst.Blocks) != 2 {
			t.Errorf("unexpected result: %+v", dst)
		}
	})

	t.Run("invalid out type", func(t *testing.T) {
		src := map[string]interface{}{"KeyInt": 42}
		var dst int
		err := i2s(src, dst)
		if err == nil {
			t.Error("expected error for non-pointer out")
		}
	})

	t.Run("nil data", func(t *testing.T) {
		var dst Simple
		err := i2s(nil, &dst)
		if err == nil {
			t.Error("expected error for nil data")
		}
	})

	t.Run("type mismatch", func(t *testing.T) {
		src := map[string]interface{}{"KeyInt": "not an int"}
		var dst Simple
		err := i2s(src, &dst)
		if err == nil {
			t.Error("expected error for type mismatch")
		}
	})
}

func TestDereferencePtr(t *testing.T) {
	t.Run("non-pointer", func(t *testing.T) {
		v := reflect.ValueOf(42)
		result := dereferencePtr(v)
		if !result.Equal(v) {
			t.Error("expected same value for non-pointer")
		}
	})
	t.Run("valid pointer", func(t *testing.T) {
		i := 42
		v := reflect.ValueOf(&i)
		result := dereferencePtr(v)
		if result.Int() != 42 {
			t.Errorf("expected 42, got %v", result.Int())
		}
	})
}

func TestI2SReflect(t *testing.T) {
	t.Run("interface value", func(t *testing.T) {
		var src interface{} = map[string]interface{}{"KeyInt": 42}
		var dst Simple
		err := i2sReflect(reflect.ValueOf(src), reflect.ValueOf(&dst).Elem())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dst.KeyInt != 42 {
			t.Errorf("unexpected result: %+v", dst)
		}
	})

	t.Run("unsupported kind", func(t *testing.T) {
		type unsupported struct{ f func() }
		src := unsupported{}
		var dst unsupported
		err := i2sReflect(reflect.ValueOf(src), reflect.ValueOf(&dst).Elem())
		if err == nil {
			t.Error("expected error for unsupported kind")
		}
	})
}

func TestNestedStructures(t *testing.T) {
	t.Run("deeply nested struct", func(t *testing.T) {
		type Nested struct {
			Level1 struct {
				Level2 struct {
					Value int
				}
			}
		}

		src := map[string]interface{}{
			"Level1": map[string]interface{}{
				"Level2": map[string]interface{}{
					"Value": 42,
				},
			},
		}

		var dst Nested
		err := i2s(src, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dst.Level1.Level2.Value != 42 {
			t.Errorf("unexpected result: %+v", dst)
		}
	})

	t.Run("struct with slice of structs", func(t *testing.T) {
		src := map[string]interface{}{
			"ManySimple": []interface{}{
				map[string]interface{}{
					"KeyInt":     1,
					"KeyString":  "first",
					"KeyBool":    true,
					"KeyFloat":   1.1,
					"KeyComplex": complex(1, 1),
				},
				map[string]interface{}{
					"KeyInt":     2,
					"KeyString":  "second",
					"KeyBool":    false,
					"KeyFloat":   2.2,
					"KeyComplex": complex(2, 2),
				},
			},
		}

		var dst Complex
		err := i2s(src, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dst.ManySimple) != 2 ||
			dst.ManySimple[0].KeyInt != 1 ||
			dst.ManySimple[1].KeyString != "second" {
			t.Errorf("unexpected result: %+v", dst)
		}
	})
}

func TestSliceEdgeCases(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		src := map[string]interface{}{
			"ManySimple": []interface{}{},
		}

		var dst Complex
		err := i2s(src, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dst.ManySimple) != 0 {
			t.Errorf("expected empty slice, got %v", dst.ManySimple)
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		src := map[string]interface{}{
			"ManySimple": nil,
		}

		var dst Complex
		err := i2s(src, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dst.ManySimple != nil {
			t.Errorf("expected nil slice, got %v", dst.ManySimple)
		}
	})

	t.Run("slice type mismatch", func(t *testing.T) {
		src := map[string]interface{}{
			"ManySimple": []interface{}{
				"not a struct",
			},
		}

		var dst Complex
		err := i2s(src, &dst)
		if err == nil {
			t.Error("expected error for type mismatch in slice")
		}
	})
}

func TestComplexSliceStructures(t *testing.T) {
	t.Run("slice of complex structs", func(t *testing.T) {
		src := []interface{}{
			map[string]interface{}{
				"SubSimple": map[string]interface{}{
					"KeyInt":   1,
					"KeyFloat": 1.1,
				},
				"ManySimple": []interface{}{
					map[string]interface{}{
						"KeyString": "nested",
					},
				},
				"Blocks": []interface{}{
					map[string]interface{}{"ID": 1},
				},
			},
		}

		var dst []Complex
		err := i2s(src, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dst) != 1 ||
			len(dst[0].ManySimple) != 1 ||
			dst[0].Blocks[0].ID != 1 {
			t.Errorf("unexpected result: %+v", dst)
		}
	})

	t.Run("slice of slices", func(t *testing.T) {
		src := [][]interface{}{
			{1, 2, 3},
			{4, 5, 6},
		}

		var dst [][]int
		err := i2s(src, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dst) != 2 || len(dst[0]) != 3 || dst[1][2] != 6 {
			t.Errorf("unexpected result: %v", dst)
		}
	})
}

func TestPointerHandling(t *testing.T) {
	t.Run("pointer fields", func(t *testing.T) {
		type WithPointer struct {
			PtrInt *int
			PtrStr *string
		}

		src := map[string]interface{}{
			"PtrInt": 42,
			"PtrStr": "test",
		}

		var dst WithPointer
		err := i2s(src, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if *dst.PtrInt != 42 || *dst.PtrStr != "test" {
			t.Errorf("unexpected result: %+v", dst)
		}
	})

	t.Run("nil pointer fields", func(t *testing.T) {
		type WithPointer struct {
			PtrInt *int
		}

		src := map[string]interface{}{
			"PtrInt": nil,
		}

		var dst WithPointer
		err := i2s(src, &dst)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dst.PtrInt != nil {
			t.Errorf("expected nil pointer, got %v", dst.PtrInt)
		}
	})
}
