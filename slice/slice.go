package slice

import (
	"encoding/json"
	"errors"
	"reflect"
)

var (
	ErrNotSlice    = errors.New("Not a slice")
	ErrNotPointer  = errors.New("Not a pointer type")
	ErrNotSameType = errors.New("Not the same type")
)

// Contains takes a source Slice/Array and an element that can be slice/Array
// or a single value type of the same type as elements in source Slice/Array.
// If the source Slice/Array contains the single element or any element in
// the provided Slice/Array, it will return true.
func Contains(src interface{}, element interface{}) (bool, error) {
	sv := reflect.ValueOf(src)

	// no need to reflect a reflect.Value again, will return a struct Kind()
	var ev reflect.Value
	v, ok := element.(reflect.Value)
	if ok {
		ev = v
	} else {
		ev = reflect.ValueOf(element)
	}

	if !isSlice(sv) {
		return false, ErrNotSlice
	}

	if ev.Kind() == reflect.Slice || ev.Kind() == reflect.Array {
		for i := 0; i < sv.Len(); i++ {
			ok, e := Contains(element, sv.Index(i))
			if e != nil {
				return false, e
			}
			if ok {
				return true, nil
			}
		}
	} else {
		for i := 0; i < sv.Len(); i++ {
			if reflect.DeepEqual(sv.Index(i).Interface(), ev.Interface()) {
				return true, nil
			}
		}
	}
	return false, nil
}

func shortest(src interface{}) (dst interface{}, err error) {
	sv := reflect.ValueOf(src)

	if !isSlice(sv) {
		return dst, ErrNotSlice
	}

	for i := 0; i < sv.Len(); i++ {
		if dst == nil {
			dst = sv.Index(i).Interface()
			continue
		}
		if sv.Index(i).Kind() == reflect.String {
			if sv.Index(i).Len() < len(reflect.ValueOf(dst).String()) {
				dst = sv.Index(i).Interface()
			}
		}
		if sv.Index(i).Kind() == reflect.Int {
			if sv.Index(i).Int() < reflect.ValueOf(dst).Int() {
				dst = sv.Index(i).Interface()
			}
		}
	}
	return dst, err
}

// ShortestString find the shortest string in string slice
func ShortestString(src []string) (string, error) {
	s, e := shortest(src)
	return s.(string), e
}

// Remove takes a pointer to slice as source and an element that
// can be slice or single value type of the same type as the
// elements in the source slice.
// It will remove the single element or elements in the provided
// slice from the source slice
func Remove(src interface{}, element interface{}) error {
	sv := reflect.ValueOf(src)
	// no need to reflect a reflect.Value again, will return a struct Kind()
	var ev reflect.Value
	v, ok := element.(reflect.Value)
	if ok {
		ev = v
	} else {
		ev = reflect.ValueOf(element)
	}

	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	} else {
		return ErrNotPointer
	}

	if !isSlice(sv) {
		return ErrNotSlice
	}

	if isSlice(ev) {
		for i := 0; i < ev.Len(); i++ {
			e := Remove(src, ev.Index(i))
			if e != nil {
				return e
			}
		}
	} else {
		if sv.Type().Elem().Kind() == ev.Kind() {
			idx := []int{}
			for i := 0; i < sv.Len(); i++ {
				if reflect.DeepEqual(sv.Index(i).Interface(), ev.Interface()) {
					idx = append(idx, i)
				}
			}
			tmp := removeFromSlice(idx, sv)
			sv.Set(tmp)
			return nil
		}
		return ErrNotSameType
	}
	return nil
}

// Unique remove the duplicated element from a slice
func Unique(src interface{}) error {
	sv := reflect.ValueOf(src)
	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	} else {
		return ErrNotPointer
	}

	if !isSlice(sv) {
		return ErrNotSlice
	}

	m := make(map[interface{}]struct{})
	idx := []int{}

	for i := 0; i < sv.Len(); i++ {
		k := genKey(sv.Index(i))
		if _, ok := m[k]; ok {
			idx = append(idx, i)
		} else {
			m[k] = struct{}{}
		}
	}

	tmp := removeFromSlice(idx, sv)
	sv.Set(tmp)

	return nil
}

// Intersect find the common piece of two slice
func Intersect(src interface{}, dst interface{}) error {
	sv := reflect.ValueOf(src)
	dv := reflect.ValueOf(dst)

	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	} else {
		return ErrNotPointer
	}

	for _, v := range []reflect.Value{sv, dv} {
		if !isSlice(v) {
			return ErrNotSlice
		}
	}

	m := make(map[interface{}]struct{})
	idx := []int{}

	for i := 0; i < dv.Len(); i++ {
		m[genKey(dv.Index(i))] = struct{}{}
	}

	for j := 0; j < sv.Len(); j++ {
		if _, ok := m[genKey(sv.Index(j))]; !ok {
			idx = append(idx, j)
		}
	}

	tmp := removeFromSlice(idx, sv)
	sv.Set(tmp)

	return nil
}

// Concat append a slice or an element to the existing slice
func Concat(src interface{}, dst interface{}) error {
	sv := reflect.ValueOf(src)
	dv := reflect.ValueOf(dst)

	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	} else {
		return ErrNotPointer
	}

	if !isSlice(sv) {
		return ErrNotSlice
	}

	m := make(map[interface{}]struct{})

	for i := 0; i < sv.Len(); i++ {
		m[genKey(sv.Index(i))] = struct{}{}
	}

	if dv.Kind() == reflect.Slice {
		for j := 0; j < dv.Len(); j++ {
			if _, ok := m[genKey(dv.Index(j))]; !ok {
				sv.Set(reflect.Append(sv, dv.Index(j)))
			}
		}
	} else {
		if sv.Type().Elem().Kind() == dv.Kind() {
			if _, ok := m[genKey(dv)]; !ok {
				sv.Set(reflect.Append(sv, dv))
			}
			return nil
		}
		return ErrNotSameType
	}
	return nil
}

// Replace replace the old element in slice with the new one
func Replace(src, old, new interface{}) error {
	sv := reflect.ValueOf(src)
	nv := reflect.ValueOf(new)
	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	} else {
		return ErrNotPointer
	}

	var ov reflect.Value
	v, ok := old.(reflect.Value)
	if ok {
		ov = v
	} else {
		ov = reflect.ValueOf(old)
	}

	if sv.Type().Elem().Kind() != nv.Kind() {
		return ErrNotSameType
	}

	for i := 0; i < sv.Len(); i++ {
		if reflect.DeepEqual(sv.Index(i).Interface(), ov.Interface()) {
			sv.Index(i).Set(nv)
		}
	}

	return nil
}

// Flatten flatten slice of slices to one depth slice
func Flatten(slice interface{}) (interface{}, error) {
	sv := reflect.ValueOf(slice)
	if sv.Kind() == reflect.Slice || sv.Kind() == reflect.Array {
		length := 0
		pos := 0
		for i := 0; i < sv.Len(); i++ {
			in := sv.Index(i)
			if in.Kind() == reflect.Slice || in.Kind() == reflect.Array {
				length += in.Len()
			} else {
				// not slice of slice, just return the original slice.
				return slice, nil
			}
		}
		s := reflect.MakeSlice(reflect.SliceOf(sv.Index(0).Index(0).Type()), length, length)
		for i := 0; i < sv.Len(); i++ {
			in := sv.Index(i)
			for j := 0; j < in.Len(); j++ {
				v := in.Index(j)
				if s.Index(pos).CanSet() {
					s.Index(pos).Set(v)
				}
				pos++
			}
		}
		return s.Interface(), nil
	}
	// you can even flatten non-slice/array stuff
	return slice, nil
}

//genKey generate map key
// currently support all fully comparable type and struct
func genKey(v reflect.Value) interface{} {
	k := v.Interface()

	if v.Kind() == reflect.Struct {
		b, _ := json.Marshal(k)
		k = reflect.ValueOf(string(b)).Interface()
	}

	return k
}

func isSlice(v reflect.Value) bool {
	if v.Kind() == reflect.Slice {
		return true
	}
	return false
}

func removeFromSlice(idx []int, v reflect.Value) reflect.Value {
	tmp := reflect.MakeSlice(v.Type(), v.Len()-len(idx), v.Cap()-len(idx))
	n := 0
	for i := 0; i < v.Len(); i++ {
		has := false
		for _, j := range idx {
			if j == i {
				has = true
				n++
				break
			}
		}
		if !has {
			tmp.Index(i - n).Set(v.Index(i))
		}
	}
	return tmp
}
