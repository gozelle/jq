package jq_test

import (
	"fmt"
	"github.com/gozelle/jq"
	"log"
)

type moduleLoader struct{}

func (*moduleLoader) LoadModule(name string) (*jq.Query, error) {
	switch name {
	case "module1":
		return jq.Parse(`
			module { name: "module1", test: 42 };
			import "module2" as foo;
			def f: foo::f;
		`)
	case "module2":
		return jq.Parse(`
			def f: .foo;
		`)
	}
	return nil, fmt.Errorf("module not found: %q", name)
}

func ExampleWithModuleLoader() {
	query, err := jq.Parse(`
		import "module1" as m;
		m::f
	`)
	if err != nil {
		log.Fatalln(err)
	}
	code, err := jq.Compile(
		query,
		jq.WithModuleLoader(&moduleLoader{}),
	)
	if err != nil {
		log.Fatalln(err)
	}
	input := map[string]interface{}{"foo": 42}
	iter := code.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}
		fmt.Printf("%#v\n", v)
	}

	// Output:
	// 42
}
