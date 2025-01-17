package jq_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gozelle/jq"
	"reflect"
	"testing"
)

func TestWithModuleLoaderError(t *testing.T) {
	query, err := jq.Parse(`
		import "module1" as m;
		m::f
	`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = jq.Compile(query)
	if got, expected := err.Error(), `cannot load module: "module1"`; got != expected {
		t.Errorf("expected: %v, got: %v", expected, got)
	}

	query, err = jq.Parse("modulemeta")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(query)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run("m")
	v, ok := iter.Next()
	if !ok {
		t.Fatal("should emit an error but got no output")
	}
	if err, ok := v.(error); ok {
		if expected := `cannot load module: "m"`; err.Error() != expected {
			t.Errorf("expected: %v, got: %v", expected, err)
		}
	} else {
		t.Errorf("should emit an error but got: %v", v)
	}
	v, ok = iter.Next()
	if ok {
		t.Errorf("should not emit a value but got: %v", v)
	}
}

func TestWithModuleLoader_modulemeta(t *testing.T) {
	query, err := jq.Parse(`
		"module1" | modulemeta
	`)
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(
		query,
		jq.WithModuleLoader(&moduleLoader{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	for {
		got, ok := iter.Next()
		if !ok {
			break
		}
		if expected := map[string]interface{}{
			"deps": []interface{}{
				map[string]interface{}{
					"relpath": "module2",
					"as":      "foo",
					"is_data": false,
				},
			},
			"name": "module1",
			"test": 42,
		}; !reflect.DeepEqual(got, expected) {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	}
}

type moduleLoaderJSON struct{}

func (*moduleLoaderJSON) LoadJSON(name string) (interface{}, error) {
	if name == "module1" {
		return []interface{}{1.0, 42, json.Number("123")}, nil
	}
	return nil, fmt.Errorf("module not found: %q", name)
}

func TestWithModuleLoader_JSON(t *testing.T) {
	query, err := jq.Parse(`
		import "module1" as $m;
		[$m, $m[1]*$m[2]]
	`)
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(
		query,
		jq.WithModuleLoader(&moduleLoaderJSON{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	for {
		got, ok := iter.Next()
		if !ok {
			break
		}
		if expected := []interface{}{[]interface{}{1.0, 42, 123}, 5166}; !reflect.DeepEqual(got, expected) {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	}
}

type moduleLoaderInitModules struct{}

func (*moduleLoaderInitModules) LoadInitModules() ([]*jq.Query, error) {
	query, err := jq.Parse(`
		def f: 42;
		def g: f * f;
	`)
	if err != nil {
		return nil, err
	}
	return []*jq.Query{query}, nil
}

func TestWithModuleLoader_LoadInitModules(t *testing.T) {
	query, err := jq.Parse("g")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(
		query,
		jq.WithModuleLoader(&moduleLoaderInitModules{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	for {
		got, ok := iter.Next()
		if !ok {
			break
		}
		if expected := 42 * 42; !reflect.DeepEqual(got, expected) {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	}
}

func TestWithEnvironLoader(t *testing.T) {
	query, err := jq.Parse("env")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(
		query,
		jq.WithEnvironLoader(func() []string {
			return []string{"foo=42", "bar=128", "baz", "qux=", "=0"}
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	for {
		got, ok := iter.Next()
		if !ok {
			break
		}
		expected := map[string]interface{}{"foo": "42", "bar": "128", "qux": ""}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	}
}

func TestWithEnvironLoaderEmpty(t *testing.T) {
	query, err := jq.Parse("env")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(query)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	for {
		got, ok := iter.Next()
		if !ok {
			break
		}
		if expected := map[string]interface{}{}; !reflect.DeepEqual(got, expected) {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	}
}

func TestWithVariablesError0(t *testing.T) {
	query, err := jq.Parse(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"", "$", "_x", "x", "$0x", "$$", "$x ", " $x"} {
		_, err = jq.Compile(
			query,
			jq.WithVariables([]string{name}),
		)
		if err == nil {
			t.Fatalf("%q is not a valid variable name", name)
		}
		if got, expected := err.Error(), "invalid variable name: "+name; got != expected {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	}
}

func TestWithVariablesError1(t *testing.T) {
	query, err := jq.Parse(".")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(
		query,
		jq.WithVariables([]string{"$x"}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	v, ok := iter.Next()
	if !ok {
		t.Fatal("should emit an error but got no output")
	}
	if err, ok := v.(error); ok {
		if expected := "variable defined but not bound: $x"; err.Error() != expected {
			t.Errorf("expected: %v, got: %v", expected, err)
		}
	} else {
		t.Errorf("should emit an error but got: %v", v)
	}
	v, ok = iter.Next()
	if ok {
		t.Errorf("should not emit a value but got: %v", v)
	}
}

func TestWithVariablesError2(t *testing.T) {
	query, err := jq.Parse(".")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(
		query,
		jq.WithVariables([]string{"$x"}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil, 1, 2)
	v, ok := iter.Next()
	if !ok {
		t.Fatal("should emit an error but got no output")
	}
	if err, ok := v.(error); ok {
		if expected := "too many variable values provided"; err.Error() != expected {
			t.Errorf("expected: %v, got: %v", expected, err)
		}
	} else {
		t.Errorf("should emit an error but got: %v", v)
	}
	v, ok = iter.Next()
	if ok {
		t.Errorf("should not emit a value but got: %v", v)
	}
}

func TestWithFunction(t *testing.T) {
	options := []jq.CompilerOption{
		jq.WithFunction("f", 0, 0, func(x interface{}, _ []interface{}) interface{} {
			if x, ok := x.(int); ok {
				return x * 2
			}
			return fmt.Errorf("f cannot be applied to: %v", x)
		}),
		jq.WithFunction("g", 1, 1, func(x interface{}, xs []interface{}) interface{} {
			if x, ok := x.(int); ok {
				if y, ok := xs[0].(int); ok {
					return x + y
				}
			}
			return fmt.Errorf("g cannot be applied to: %v, %v", x, xs)
		}),
	}
	query, err := jq.Parse(".[] | f | g(3)")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(query, options...)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run([]interface{}{0, 1, 2, 3, 4})
	n := 0
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if expected := n*2 + 3; v != expected {
			t.Errorf("expected: %v, got: %v", expected, v)
		}
		n++
	}
	if expected := 5; n != expected {
		t.Errorf("expected: %v, got: %v", expected, n)
	}
	query, err = jq.Parse(
		`("f/0", "f/1", "g/0", "g/1") as $f | builtins | any(. == $f)`,
	)
	if err != nil {
		t.Fatal(err)
	}
	code, err = jq.Compile(query, options...)
	if err != nil {
		t.Fatal(err)
	}
	iter = code.Run(nil)
	n = 0
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if expected := n == 0 || n == 3; v != expected {
			t.Errorf("expected: %v, got: %v (n = %d)", expected, v, n)
		}
		n++
	}
	if expected := 4; n != expected {
		t.Errorf("expected: %v, got: %v", expected, n)
	}
}

func TestWithFunctionDuplicateName(t *testing.T) {
	options := []jq.CompilerOption{
		jq.WithFunction("f", 0, 0, func(x interface{}, _ []interface{}) interface{} {
			if x, ok := x.(int); ok {
				return x * 2
			}
			return fmt.Errorf("f cannot be applied to: %v", x)
		}),
		jq.WithFunction("f", 1, 1, func(x interface{}, xs []interface{}) interface{} {
			if x, ok := x.(int); ok {
				if y, ok := xs[0].(int); ok {
					return x + y
				}
			}
			return fmt.Errorf("f cannot be applied to: %v, %v", x, xs)
		}),
		jq.WithFunction("f", 0, 0, func(x interface{}, _ []interface{}) interface{} {
			if x, ok := x.(int); ok {
				return x * 4
			}
			return fmt.Errorf("f cannot be applied to: %v", x)
		}),
		jq.WithFunction("f", 2, 2, func(x interface{}, xs []interface{}) interface{} {
			if x, ok := x.(int); ok {
				if y, ok := xs[0].(int); ok {
					if z, ok := xs[1].(int); ok {
						return x * y * z
					}
				}
			}
			return fmt.Errorf("f cannot be applied to: %v, %v", x, xs)
		}),
	}
	query, err := jq.Parse(".[] | f | f(3) | f(4; 5)")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(query, options...)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run([]interface{}{0, 1, 2, 3, 4})
	n := 0
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if expected := (n*4 + 3) * 4 * 5; v != expected {
			t.Errorf("expected: %v, got: %v", expected, v)
		}
		n++
	}
	if expected := 5; n != expected {
		t.Errorf("expected: %v, got: %v", expected, n)
	}
	query, err = jq.Parse(
		`("f/0", "f/1", "f/2", "f/3") as $f | builtins | any(. == $f)`,
	)
	if err != nil {
		t.Fatal(err)
	}
	code, err = jq.Compile(query, options...)
	if err != nil {
		t.Fatal(err)
	}
	iter = code.Run(nil)
	n = 0
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if expected := n < 3; v != expected {
			t.Errorf("expected: %v, got: %v (n = %d)", expected, v, n)
		}
		n++
	}
	if expected := 4; n != expected {
		t.Errorf("expected: %v, got: %v", expected, n)
	}
}

func TestWithFunctionMultipleArities(t *testing.T) {
	options := []jq.CompilerOption{
		jq.WithFunction("f", 0, 4, func(x interface{}, xs []interface{}) interface{} {
			if x, ok := x.(int); ok {
				x *= 2
				for _, y := range xs {
					if y, ok := y.(int); ok {
						x += y
					}
				}
				return x
			}
			return fmt.Errorf("f cannot be applied to: %v, %v", x, xs)
		}),
		jq.WithFunction("f", 2, 3, func(x interface{}, xs []interface{}) interface{} {
			if x, ok := x.(int); ok {
				for _, y := range xs {
					if y, ok := y.(int); ok {
						x *= y
					}
				}
				return x
			}
			return fmt.Errorf("f cannot be applied to: %v, %v", x, xs)
		}),
		jq.WithFunction("g", 0, 30, func(x interface{}, xs []interface{}) interface{} {
			return nil
		}),
	}
	query, err := jq.Parse(".[] | f | f(1) | f(2; 3) | f(4; 5; 6) | f(7; 8; 9; 10)")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(query, options...)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run([]interface{}{0, 1, 2, 3, 4})
	n := 0
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if expected := (((n*2*2+1)*2*3)*4*5*6)*2 + 7 + 8 + 9 + 10; v != expected {
			t.Errorf("expected: %v, got: %v", expected, v)
		}
		n++
	}
	if expected := 5; n != expected {
		t.Errorf("expected: %v, got: %v", expected, n)
	}
	query, err = jq.Parse(
		`("f/0", "f/1", "f/2", "f/3", "f/4", "f/5") as $f | builtins | any(. == $f)`,
	)
	if err != nil {
		t.Fatal(err)
	}
	code, err = jq.Compile(query, options...)
	if err != nil {
		t.Fatal(err)
	}
	iter = code.Run(nil)
	n = 0
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if expected := n < 5; v != expected {
			t.Errorf("expected: %v, got: %v (n = %d)", expected, v, n)
		}
		n++
	}
	if expected := 6; n != expected {
		t.Errorf("expected: %v, got: %v", expected, n)
	}
}

type valueError struct {
	v interface{}
}

func (err *valueError) Error() string {
	return "error: " + fmt.Sprint(err.v)
}

func (err *valueError) Value() interface{} {
	return err.v
}

func TestWithFunctionValueError(t *testing.T) {
	expected := map[string]interface{}{"foo": 42}
	query, err := jq.Parse("try f catch .")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(query,
		jq.WithFunction("f", 0, 0, func(x interface{}, _ []interface{}) interface{} {
			return &valueError{expected}
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if !reflect.DeepEqual(v, expected) {
			t.Errorf("expected: %v, got: %v", expected, v)
		}
	}
}

func TestWithFunctionCompileArgsError(t *testing.T) {
	query, err := jq.Parse("f(g)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = jq.Compile(query,
		jq.WithFunction("f", 0, 1, func(interface{}, []interface{}) interface{} {
			return 0
		}),
	)
	if got, expected := err.Error(), "function not defined: g/0"; got != expected {
		t.Errorf("expected: %v, got: %v", expected, got)
	}
}

func TestWithFunctionArityError(t *testing.T) {
	query, err := jq.Parse("f")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ min, max int }{{3, 2}, {-1, 3}, {0, 31}, {-1, 31}} {
		func() {
			defer func() {
				expected := fmt.Sprintf(`invalid arity for "f": %d, %d`, tc.min, tc.max)
				if got := recover(); got != expected {
					t.Errorf("expected: %v, got: %v", expected, got)
				}
			}()
			t.Fatal(jq.Compile(query,
				jq.WithFunction("f", tc.min, tc.max, func(interface{}, []interface{}) interface{} {
					return 0
				}),
			))
		}()
	}
}

func TestWithIterFunction(t *testing.T) {
	query, err := jq.Parse("f * g(5; 10), h, 10")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(query,
		jq.WithIterFunction("f", 0, 0, func(interface{}, []interface{}) jq.Iter {
			return jq.NewIter(1, 2, 3)
		}),
		jq.WithIterFunction("g", 2, 2, func(_ interface{}, xs []interface{}) jq.Iter {
			if x, ok := xs[0].(int); ok {
				if y, ok := xs[1].(int); ok {
					return &rangeIter{x, y}
				}
			}
			return jq.NewIter(fmt.Errorf("g cannot be applied to: %v", xs))
		}),
		jq.WithIterFunction("h", 0, 0, func(interface{}, []interface{}) jq.Iter {
			return jq.NewIter()
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	n := 0
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if expected := (1 + n%3) * (5 + n/3); v != expected {
			t.Errorf("expected: %v, got: %v", expected, v)
		}
		n++
	}
	if expected := 16; n != expected {
		t.Errorf("expected: %v, got: %v", expected, n)
	}
}

func TestWithIterFunctionError(t *testing.T) {
	query, err := jq.Parse("range(3) * (f, 0), f")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(query,
		jq.WithIterFunction("f", 0, 0, func(interface{}, []interface{}) jq.Iter {
			return jq.NewIter(1, errors.New("error"), 3)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	n := 0
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		switch n {
		case 0, 1, 2:
			if expected := n; v != expected {
				t.Errorf("expected: %v, got: %v", expected, v)
			}
		case 3, 5:
			if expected := "error"; v.(error).Error() != expected {
				t.Errorf("expected: %v, got: %v", expected, err)
			}
		default:
			if expected := n - 3; v != expected {
				t.Errorf("expected: %v, got: %v", expected, v)
			}
		}
		n++
	}
	if expected := 7; n != expected {
		t.Errorf("expected: %v, got: %v", expected, n)
	}
}

func TestWithIterFunctionPathIndexing(t *testing.T) {
	query, err := jq.Parse(".[f] = 1")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(query,
		jq.WithIterFunction("f", 0, 0, func(interface{}, []interface{}) jq.Iter {
			return jq.NewIter(0, 1, 2)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if expected := []interface{}{1, 1, 1}; !reflect.DeepEqual(v, expected) {
			t.Errorf("expected: %v, got: %v", expected, v)
		}
	}
}

func TestWithIterFunctionPathInputValue(t *testing.T) {
	query, err := jq.Parse("{x: 0} | (f|.x) = 1")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(query,
		jq.WithIterFunction("f", 0, 0, func(v interface{}, _ []interface{}) jq.Iter {
			return jq.NewIter(v, v, v)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if expected := map[string]interface{}{"x": 1}; !reflect.DeepEqual(v, expected) {
			t.Errorf("expected: %v, got: %v", expected, v)
		}
	}
}

func TestWithIterFunctionPathEmpty(t *testing.T) {
	query, err := jq.Parse("{x: 0} | (f|.x) = 1")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(query,
		jq.WithIterFunction("f", 0, 0, func(interface{}, []interface{}) jq.Iter {
			return jq.NewIter()
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if expected := map[string]interface{}{"x": 0}; !reflect.DeepEqual(v, expected) {
			t.Errorf("expected: %v, got: %v", expected, v)
		}
	}
}

func TestWithIterFunctionInvalidPathError(t *testing.T) {
	query, err := jq.Parse("{x: 0} | (f|.x) = 1")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(query,
		jq.WithIterFunction("f", 0, 0, func(interface{}, []interface{}) jq.Iter {
			return jq.NewIter(map[string]interface{}{"x": 1})
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	v, ok := iter.Next()
	if !ok {
		t.Fatal("should emit an error but got no output")
	}
	if err, ok := v.(error); ok {
		if expected := `invalid path against: object ({"x":1})`; err.Error() != expected {
			t.Errorf("expected: %v, got: %v", expected, err)
		}
	} else {
		t.Errorf("should emit an error but got: %v", v)
	}
}

func TestWithIterFunctionPathError(t *testing.T) {
	query, err := jq.Parse("{x: 0} | (f|.x) = 1")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(query,
		jq.WithIterFunction("f", 0, 0, func(interface{}, []interface{}) jq.Iter {
			return jq.NewIter(errors.New("error"))
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	v, ok := iter.Next()
	if !ok {
		t.Fatal("should emit an error but got no output")
	}
	if err, ok := v.(error); ok {
		if expected := "error"; err.Error() != expected {
			t.Errorf("expected: %v, got: %v", expected, err)
		}
	} else {
		t.Errorf("should emit an error but got: %v", v)
	}
}

func TestWithIterFunctionDefineError(t *testing.T) {
	query, err := jq.Parse("f")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		expected := `cannot define both iterator and non-iterator functions for "f"`
		if got := recover(); got != expected {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	}()
	t.Fatal(jq.Compile(query,
		jq.WithFunction("f", 0, 0, func(interface{}, []interface{}) interface{} {
			return 0
		}),
		jq.WithIterFunction("f", 0, 0, func(interface{}, []interface{}) jq.Iter {
			return jq.NewIter()
		}),
	))
}

type moduleLoader2 struct{}

func (*moduleLoader2) LoadModule(name string) (*jq.Query, error) {
	if name == "module1" {
		return jq.Parse(`
			def g: def h: f * 3; h * 4;
		`)
	}
	return nil, fmt.Errorf("module not found: %q", name)
}

func TestWithFunctionWithModuleLoader(t *testing.T) {
	query, err := jq.Parse(`include "module1"; .[] | g`)
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(query,
		jq.WithFunction("f", 0, 0, func(x interface{}, _ []interface{}) interface{} {
			if x, ok := x.(int); ok {
				return x * 2
			}
			return fmt.Errorf("f cannot be applied to: %v", x)
		}),
		jq.WithModuleLoader(&moduleLoader2{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run([]interface{}{0, 1, 2, 3, 4})
	n := 0
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if expected := n * 2 * 3 * 4; v != expected {
			t.Errorf("expected: %v, got: %v", expected, v)
		}
		n++
	}
	if expected := 5; n != expected {
		t.Errorf("expected: %v, got: %v", expected, n)
	}
}

func TestWithInputIter(t *testing.T) {
	query, err := jq.Parse("range(10) | input")
	if err != nil {
		t.Fatal(err)
	}
	code, err := jq.Compile(
		query,
		jq.WithInputIter(jq.NewIter(0, 1, 2, 3, 4)),
	)
	if err != nil {
		t.Fatal(err)
	}
	iter := code.Run(nil)
	n := 0
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if expected := "break"; err.Error() != expected {
				t.Errorf("expected: %v, got: %v", expected, err)
			}
		} else if v != n {
			t.Errorf("expected: %v, got: %v", n, v)
		}
		n++
	}
	if expected := 10; n != expected {
		t.Errorf("expected: %v, got: %v", expected, n)
	}
}
