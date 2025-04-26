// Package main provides utilities to decode and map generic data structures
// (such as map[string]interface{}) into strongly typed Go structs using reflection.
package main

import (
	"errors"
	"fmt"
	"math"
	"reflect"
)

// mapStructFieldsByName maps the field names of a struct to their corresponding reflect.Value.
// It returns an error if the input is not a struct or a pointer to a struct.
func mapStructFieldsByName(out reflect.Value) (map[string]reflect.Value, error) {
	if out.Kind() == reflect.Pointer {
		out = out.Elem()
	}

	if out.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", out.Kind().String())
	}

	mp := make(map[string]reflect.Value)

	for i := range out.NumField() {
		fieldName := out.Type().Field(i).Name
		mp[fieldName] = out.Field(i)
	}

	return mp, nil
}

// assignSimpleValue assigns a simple value (int, float, bool, string, complex) from src to dst,
// handling type conversion where appropriate. Returns an error on incompatible types.
func assignSimpleValue(dst reflect.Value, src reflect.Value) error {
	dstType := dst.Type().Kind()
	srcType := src.Type().Kind()

	// convert source to destination type if compatible.
	switch dstType {
	case reflect.Pointer:
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return assignSimpleValue(dst.Elem(), src)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch srcType {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			dst.SetInt(src.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val := src.Uint()
			if val > math.MaxInt64 {
				return errors.New("value too large to convert uint to int64")
			}
			dst.SetInt(int64(val))
		case reflect.Float32, reflect.Float64:
			dst.SetInt(int64(src.Float()))
		default:
			return fmt.Errorf("cannot assign value of type %s to int field %q", srcType, dst.Type().Name())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch srcType {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val := src.Int()
			if val < 0 {
				return errors.New("cannot convert negative int to uint")
			}
			dst.SetUint(uint64(val))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			dst.SetUint(src.Uint())
		case reflect.Float32, reflect.Float64:
			dst.SetUint(uint64(src.Float()))
		default:
			return fmt.Errorf("cannot assign value of type %s to uint field %q", srcType, dst.Type().Name())
		}
	case reflect.Float32, reflect.Float64:
		switch srcType {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			dst.SetFloat(float64(src.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			dst.SetFloat(float64(src.Uint()))
		case reflect.Float32, reflect.Float64:
			dst.SetFloat(src.Float())
		default:
			return fmt.Errorf("cannot assign value of type %s to float field %q", srcType, dst.Type().Name())
		}
	case reflect.Complex64, reflect.Complex128:
		if srcType == reflect.Complex64 || srcType == reflect.Complex128 {
			dst.SetComplex(src.Complex())
		} else {
			return fmt.Errorf("cannot assign value of type %s to complex field %q", srcType, dst.Type().Name())
		}
	case reflect.Bool:
		if srcType == reflect.Bool {
			dst.SetBool(src.Bool())
		} else {
			return fmt.Errorf("cannot assign value of type %s to bool field %q", srcType, dst.Type().Name())
		}
	case reflect.String:
		if srcType == reflect.String {
			dst.SetString(src.String())
		} else {
			return fmt.Errorf("cannot assign value of type %s to string field %q", srcType, dst.Type().Name())
		}
	default:
		return fmt.Errorf("unsupported dst type: %s", dstType)
	}
	return nil
}

// checkIfArrayOrSlice checks whether a reflect.Value is an array or a slice.
func checkIfArrayOrSlice(val reflect.Value) bool {
	kind := val.Kind()
	return kind == reflect.Array || kind == reflect.Slice
}

// allocateAndFillSlice creates a new slice of the same type as dst, fills it by recursively copying
// elements from src, and sets it to dst. Returns an error if types are incompatible.
func allocateAndFillSlice(dst reflect.Value, src reflect.Value) error {
	if !checkIfArrayOrSlice(dst) {
		return errors.New("dst is not array or slice")
	}
	if !checkIfArrayOrSlice(src) {
		return errors.New("src is not array or slice")
	}

	dstElemType := dst.Type().Elem()
	newDst := reflect.MakeSlice(dst.Type(), src.Len(), src.Len())

	for i := range src.Len() {
		srcElem := src.Index(i)
		dstElem := reflect.New(dstElemType).Elem()

		if err := i2sReflect(srcElem, dstElem); err != nil {
			return fmt.Errorf("element %d conversion failed: %w", i, err)
		}

		newDst.Index(i).Set(dstElem)
	}

	dst.Set(newDst)
	return nil
}

// assignArraySliceValue assigns values from a source slice or array to a destination slice or array.
// It handles deep copying of elements. Returns an error on failure.
func assignArraySliceValue(dst reflect.Value, src reflect.Value) error {
	if !checkIfArrayOrSlice(dst) {
		return errors.New("dst is not array/slice")
	}
	if !checkIfArrayOrSlice(src) {
		return errors.New("src is not array/lice")
	}

	err := allocateAndFillSlice(dst, src)
	if err != nil {
		return err
	}

	return nil
}

// mapKeyType returns the kind of the keys used in a map reflect.Value.
// Returns an error if the input is not a map.
func mapKeyType(data reflect.Value) (reflect.Kind, error) {
	if data.Kind() != reflect.Map {
		return reflect.Invalid, fmt.Errorf("expected map, got %s", data.Kind().String())
	}
	return data.Type().Key().Kind(), nil
}

// assignMap maps key-value pairs from a map[string]interface{} to fields of a struct.
// Fields not present in the struct are ignored.
func assignMap(data reflect.Value, out reflect.Value) error {
	if data.IsNil() {
		return nil
	}

	fieldsMap, err := mapStructFieldsByName(out)
	if err != nil {
		return err
	}

	mapKeyType, err := mapKeyType(data)
	if err != nil {
		return err
	}
	if mapKeyType != reflect.String {
		return fmt.Errorf("expected map with string key, got %s", mapKeyType.String())
	}

	for _, key := range data.MapKeys() {
		value := data.MapIndex(key)
		outField, ok := fieldsMap[key.String()]
		if !ok {
			continue
		}

		err = i2sReflect(value, outField)
		if err != nil {
			return err
		}
	}

	return nil
}

// dereferencePtr follows pointer or interface chains to get the underlying non-pointer, non-interface value.
// If the value is nil, it returns as-is.
func dereferencePtr(out reflect.Value) reflect.Value {
	for {
		switch out.Kind() {
		case reflect.Pointer, reflect.Interface:
			if out.IsNil() {
				return out
			}
			out = out.Elem()
		default:
			return out
		}
	}
}

// i2sReflect recursively assigns data from a reflect.Value into a target reflect.Value.
// Handles basic types, maps, slices/arrays, and interfaces.
func i2sReflect(data reflect.Value, out reflect.Value) error {
	out = dereferencePtr(out)
	switch data.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.Bool,
		reflect.String:
		// assign simple types.
		err := assignSimpleValue(out, data)
		if err != nil {
			return fmt.Errorf("assigning %s failed: %w", data.Type().Name(), err)
		}
		return nil
	case reflect.Map:
		return assignMap(data, out)
	case reflect.Array, reflect.Slice:
		return assignArraySliceValue(out, data)
	case reflect.Interface:
		// unwrap interface and retry.
		if data.IsNil() {
			return nil
		}
		data = dereferencePtr(data)
		return i2sReflect(data, out)
	case reflect.Invalid:
		return nil
	default:
		return fmt.Errorf("unsupported kind: %s", data.Kind())
	}
}

// i2s is the top-level function that converts a generic data structure (like a map or slice)
// into a strongly typed struct. `out` must be a pointer to the struct.
func i2s(data interface{}, out interface{}) error {
	if data == nil {
		return errors.New("data cannot be nil")
	}

	dataVal := reflect.ValueOf(data)
	outVal := reflect.ValueOf(out)

	if outVal.Kind() != reflect.Ptr {
		return fmt.Errorf("out must be a pointer, got %s", reflect.TypeOf(out).Kind())
	}

	return i2sReflect(dataVal, outVal)
}

// Decoder is a struct used to perform decoding of generic data into typed structs.
type Decoder struct{}

// NewDecoder creates a new instance of Decoder.
func NewDecoder() *Decoder {
	return &Decoder{}
}

// Decode decodes the provided generic data into the given output struct pointer.
// It returns an error if the decoding fails.
func (d *Decoder) Decode(data interface{}, out interface{}) error {
	return i2s(data, out)
}
