// Copyright 2021 Maxime THIEBAUT. All rights reserved.
// Use of this source code is governed by EUPL-1.2
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
)

// KeyError is an error whose key can be recursively set.
type KeyError interface {
	error
	Key() string
	// From prepends the KeyError's key with the provided key.
	From(key string)
}

// ConfigurationError is the base error implementing KeyError.
type ConfigurationError struct {
	Keys string
}

func (e *ConfigurationError) Key() string {
	return e.Keys
}

func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("configuration key %#v error", e.Key())
}

func (e *ConfigurationError) From(key string) {
	e.Keys = key + "." + e.Keys
}

type ErrNoSuchKey struct {
	*ConfigurationError
}

func (e *ErrNoSuchKey) Error() string {
	return fmt.Sprintf("no such %#v configuration key", e.Key())
}

type ErrUnhandledKind struct {
	*ConfigurationError
	Kind string
}

func (e *ErrUnhandledKind) Error() string {
	return fmt.Sprintf("configuration key %#v has an undhandled kind %#v", e.Key(), e.Kind)
}

type ErrIncompatibleType struct {
	*ConfigurationError
	Type string
}

func (e *ErrIncompatibleType) Error() string {
	return fmt.Sprintf("configuration key %#v has an incompatible kind %#v", e.Key(), e.Type)
}
