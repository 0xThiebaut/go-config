// Copyright 2021 Maxime THIEBAUT. All rights reserved.
// Use of this source code is governed by EUPL-1.2
// license that can be found in the LICENSE file.

package config

import (
	"reflect"
	"strconv"
	"strings"
)

// Reader abstracts a readable configuration.
type Reader interface {
	Read(key string) (interface{}, error)
	ReadString(key string) (string, error)
}

// Writer abstracts a writable configuration
type Writer interface {
	Write(key string, v interface{}) error
}

// ReadWriter abstracts a readable and writable configuration.
type ReadWriter interface {
	Reader
	Writer
}

// New creates a new ReadWriter configuration linked to the interface v.
func New(v interface{}) ReadWriter {
	return &config{Data: v}
}

// config is a recursive ReadWriter implementation
type config struct {
	Data interface{}
}

// Write sets a key's value.
func (c *config) Write(key string, value interface{}) error {
	d := reflect.ValueOf(c.Data)
	k := strings.Split(key, ".")
	v, err := c.write(k, d, value)
	if err != nil {
		return err
	}
	c.Data = v.Interface()
	return nil
}

// write recursively sets a key's value. It provides the inspected element and returns the modified element.
// By providing a modified element, write introduces support for value-passed parameters in addition to reference-passed ones.
func (c *config) write(key []string, element reflect.Value, value interface{}) (reflect.Value, KeyError) {
	if len(key) == 0 {
		return reflect.ValueOf(value), nil
	}

	switch k := element.Kind(); k {
	case reflect.Interface:
		e := element.Elem()
		e, err := c.write(key, e, value)
		if err != nil {
			return element, err
		}
		return reflect.ValueOf(e.Interface()), nil
	case reflect.Ptr:
		e := element.Elem()
		e, err := c.write(key, e, value)
		if err != nil {
			return element, err
		}
		if e.CanAddr() {
			return e.Addr(), nil
		}
		p := reflect.New(e.Type())
		p.Elem().Set(e)
		return p, nil
	case reflect.Struct:
		// Consume one key level
		name := key[0]
		key = key[1:]
		// Loop the elements
		t := element.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if strings.EqualFold(name, f.Name) {
				e := element.Field(i)
				v, err := c.write(key, e, value)
				if err != nil {
					err.From(name)
					return element, err
				}
				if !v.CanConvert(f.Type) {
					return element, &ErrIncompatibleType{Type: f.Type.String(), ConfigurationError: &ConfigurationError{name}}
				}
				if !e.CanSet() {
					n := reflect.Indirect(reflect.New(t))
					n.Set(element)
					element = n
					e = n.Field(i)
				}
				e.Set(v.Convert(f.Type))
				return element, nil
			}
		}
		return element, &ErrNoSuchKey{&ConfigurationError{name}}
	case reflect.Map:
		// Consume one key level
		name := key[0]
		key = key[1:]
		// Ensure the map is not nil
		if element.IsNil() {
			element = reflect.MakeMap(element.Type())
		}
		// Loop the elements
		i := element.MapRange()
		for i.Next() {
			// Find a matching key
			if strings.EqualFold(name, i.Key().String()) {
				// Continue recursing on the value
				e, err := c.write(key, i.Value(), value)
				if err != nil {
					err.From(name)
					return element, err
				}
				// Update the map
				element.SetMapIndex(i.Key(), e)
				return element, nil
			}
		}
		// Create a new value otherwise
		t := element.Type().Elem()
		e := reflect.Indirect(reflect.New(t))
		e, err := c.write(key, e, value)
		if err != nil {
			err.From(name)
			return element, err
		}
		if !e.CanConvert(t) {
			return element, &ErrIncompatibleType{Type: t.String(), ConfigurationError: &ConfigurationError{name}}
		}
		element.SetMapIndex(reflect.ValueOf(name), e.Convert(t))
		return element, nil
	default:
		name := key[0]
		return element, &ErrUnhandledKind{Kind: k.String(), ConfigurationError: &ConfigurationError{name}}
	}
}

// Read gets a key's value.
func (c *config) Read(key string) (interface{}, error) {
	d := reflect.ValueOf(c.Data)
	k := strings.Split(key, ".")
	return c.read(k, d)
}

// read recursively gets a key's value. It provides the inspected element and returns the final value.
func (c *config) read(key []string, element reflect.Value) (interface{}, KeyError) {
	if len(key) == 0 {
		return element.Interface(), nil
	}

	switch k := element.Kind(); k {
	case reflect.Interface:
		e := element.Elem()
		return c.read(key, e)
	case reflect.Ptr:
		e := element.Elem()
		return c.read(key, e)
	case reflect.Struct:
		// Consume one key level
		name := key[0]
		key = key[1:]
		// Loop the elements
		t := element.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if strings.EqualFold(name, f.Name) {
				e := element.Field(i)
				v, err := c.read(key, e)
				if err != nil {
					err.From(name)
					return v, err
				}
				return v, nil
			}
		}
		return nil, &ErrNoSuchKey{&ConfigurationError{name}}
	case reflect.Map:
		// Consume one key level
		name := key[0]
		key = key[1:]
		// Ensure the map is not nil
		if element.IsNil() {
			return nil, &ErrNoSuchKey{&ConfigurationError{name}}
		}
		// Loop the elements
		i := element.MapRange()
		for i.Next() {
			// Find a matching key
			if strings.EqualFold(name, i.Key().String()) {
				// Continue recursing on the value
				v, err := c.read(key, i.Value())
				if err != nil {
					err.From(name)
					return v, err
				}
				return v, nil
			}
		}
		return nil, &ErrNoSuchKey{&ConfigurationError{name}}
	default:
		name := key[0]
		return element, &ErrUnhandledKind{Kind: k.String(), ConfigurationError: &ConfigurationError{name}}
	}
}

// ReadString behaves like Read with additional conversion taking place.
func (c *config) ReadString(key string) (string, error) {
	v, err := c.Read(key)
	if err != nil {
		return "", err
	}
	val := reflect.ValueOf(v)
	switch k := val.Kind(); k {
	case reflect.String:
		return val.String(), err
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10), nil
	case reflect.Float32:
		return strconv.FormatFloat(val.Float(), 'g', -1, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'g', -1, 64), nil
	case reflect.Complex64:
		return strconv.FormatComplex(val.Complex(), 'g', -1, 64), nil
	case reflect.Complex128:
		return strconv.FormatComplex(val.Complex(), 'g', -1, 128), nil
	case reflect.Bool:
		return strconv.FormatBool(val.Bool()), nil
	default:
		// Attempt conversion
		t := reflect.TypeOf("")
		if val.CanConvert(t) {
			return val.Convert(t).String(), nil
		}
		// Error otherwise
		return "", &ErrUnhandledKind{Kind: k.String(), ConfigurationError: &ConfigurationError{key}}
	}
}

// Sub abstracts a ReadWriter sub-configuration by prefixing all keyed calls with a prefix.
//
// Sub allows for abstractions such as profiles where all `my.key` can be prefixed for example by `profiles.default`,
// resulting in the `profiles.default.my.key` key.
func Sub(rw ReadWriter, prefix string) ReadWriter {
	return &sub{RW: rw, Prefix: prefix}
}

// sub is a ReadWriter sub-configuration, prefixing all keyed calls with a prefix.
//
// sub allows for abstractions such as profiles where all `my.key` can be prefixed for example by `profiles.default`,
// resulting in the `profiles.default.my.key` key.
type sub struct {
	RW     ReadWriter
	Prefix string
}

// resolve prefixes a key with the sub prefix.
func (s *sub) resolve(key string) string {
	return s.Prefix + "." + key
}

// Read is a prefixed wrapper around the Reader.
func (s *sub) Read(key string) (interface{}, error) {
	return s.RW.Read(s.resolve(key))
}

// ReadString is a prefixed wrapper around the Reader.
func (s *sub) ReadString(key string) (string, error) {
	return s.RW.ReadString(s.resolve(key))
}

// Write is a prefixed wrapper around Writer.
func (s *sub) Write(key string, v interface{}) error {
	return s.RW.Write(s.resolve(key), v)
}
