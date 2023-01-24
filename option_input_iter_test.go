package jq_test

import (
	"fmt"
	"github.com/gozelle/jq"
	"log"
)

func ExampleWithInputIter() {
	query, err := jq.Parse("reduce inputs as $x (0; . + $x)")
	if err != nil {
		log.Fatalln(err)
	}
	code, err := jq.Compile(
		query,
		jq.WithInputIter(jq.NewIter(1, 2, 3, 4, 5)),
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
	// 15
}
