package jq_test

import (
	"fmt"
	"github.com/gozelle/jq"
	"log"
)

func ExampleWithVariables() {
	query, err := jq.Parse("$x * 100 + $y, $z")
	if err != nil {
		log.Fatalln(err)
	}
	code, err := jq.Compile(
		query,
		jq.WithVariables([]string{
			"$x", "$y", "$z",
		}),
	)
	if err != nil {
		log.Fatalln(err)
	}
	iter := code.Run(nil, 12, 42, 128)
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
	// 1242
	// 128
}
