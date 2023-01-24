package jq_test

import (
	"fmt"
	"github.com/gozelle/jq"
	"log"
)

func ExampleWithEnvironLoader() {
	query, err := jq.Parse("env | keys[]")
	if err != nil {
		log.Fatalln(err)
	}
	code, err := jq.Compile(
		query,
		jq.WithEnvironLoader(func() []string {
			return []string{"foo=42", "bar=128"}
		}),
	)
	if err != nil {
		log.Fatalln(err)
	}
	iter := code.Run(nil)
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
	// "bar"
	// "foo"
}
