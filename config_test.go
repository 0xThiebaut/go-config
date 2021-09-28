// Copyright 2021 Maxime THIEBAUT. All rights reserved.
// Use of this source code is governed by EUPL-1.2
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"testing"
)

func TestConfig_WriteStructString(t *testing.T) {
	type data struct {
		Foo string
	}
	d := data{}
	c := New(&d)
	foo := "Hello World!"
	if err := c.Write("foo", foo); err != nil {
		t.Fatal(err)
	} else if foo != d.Foo {
		t.Fatalf("expected %#v, got %#v", foo, d.Foo)
	}
}

func TestConfig_WriteStructInt(t *testing.T) {
	type data struct {
		Foo int
	}
	d := data{}
	c := New(&d)
	foo := 12345
	if err := c.Write("foo", foo); err != nil {
		t.Fatal(err)
	} else if foo != d.Foo {
		t.Fatalf("expected %#v, got %#v", foo, d.Foo)
	}
}

func TestConfig_WriteStructBool(t *testing.T) {
	type data struct {
		Foo bool
	}
	d := data{}
	c := New(&d)
	foo := true
	if err := c.Write("foo", foo); err != nil {
		t.Fatal(err)
	} else if foo != d.Foo {
		t.Fatalf("expected %#v, got %#v", foo, d.Foo)
	}
}

func TestConfig_WriteStructMap(t *testing.T) {
	type data struct {
		Foo map[string]string
	}
	d := data{}
	c := New(&d)
	baz := "baz"
	if err := c.Write("foo.bar", baz); err != nil {
		t.Fatal(err)
	} else if bar, ok := d.Foo["bar"]; !ok {
		t.Fatalf("expected key to be set")
	} else if baz != bar {
		t.Fatalf("expected %#v, got %#v", baz, bar)
	}
}

func TestConfig_WriteMap(t *testing.T) {
	d := map[string]string{}
	c := New(&d)
	bar := "bar"
	if err := c.Write("foo", bar); err != nil {
		t.Fatal(err)
	} else if foo, ok := d["foo"]; !ok {
		t.Fatalf("expected key to be set")
	} else if bar != foo {
		t.Fatalf("expected %#v, got %#v", bar, foo)
	}
}

func TestConfig_WriteStructIncorrectString(t *testing.T) {
	type data struct {
		Foo string
	}
	d := data{}
	c := New(&d)
	foo := "Hello World!"
	if err := c.Write("bar", foo); err == nil {
		t.Fatal("expected error bu got none")
	}
}

func TestConfig_WriteRead(t *testing.T) {
	d := map[string]string{}
	c := New(&d)
	bar := "bar"
	if err := c.Write("foo", bar); err != nil {
		t.Fatal(err)
	}
	foo, err := c.Read("foo")
	if err != nil {
		t.Fatal(err)
	}
	if s, ok := foo.(string); !ok {
		t.Fatalf("expected %T type, got %T type", s, foo)
	} else if bar != s {
		t.Fatalf("expected %#v, got %#v", bar, s)
	}
}

func ExampleConfig_ReadString() {
	type Config struct {
		My            string
		Exotic        map[string]Config
		Configuration bool
	}
	demo := &Config{
		My: "Demo",
	}
	c := New(&demo)
	if s, err := c.ReadString("my"); err == nil {
		fmt.Println(s)
	}
	// Output: Demo
}

func ExampleConfig_Write() {
	type Config struct {
		My            string
		Exotic        map[string]Config
		Configuration bool
	}
	demo := &Config{
		My: "Demo",
	}
	c := New(&demo)
	if err := c.Write("my", "Hello World!"); err == nil {
		fmt.Println(demo.My)
	}
	// Output: Hello World!
}

func ExampleConfig_ReadStringComplex() {
	type Config struct {
		My            string
		Exotic        map[string]Config
		Configuration bool
	}
	demo := &Config{
		My: "Demo",
	}
	c := New(&demo)
	if err := c.Write("exotic.exotic.exotic.exotic.my", "Success!"); err == nil {
		fmt.Println(demo.Exotic["exotic"].Exotic["exotic"].My)
	}
	// Output: Success!
}