package jq_test

import (
	"encoding/json"
	"fmt"
	"github.com/gozelle/jq"
	"log"
	"math/big"
	"strconv"
)

func toFloat(x interface{}) (float64, bool) {
	switch x := x.(type) {
	case int:
		return float64(x), true
	case float64:
		return x, true
	case *big.Int:
		f, err := strconv.ParseFloat(x.String(), 64)
		return f, err == nil
	default:
		return 0.0, false
	}
}

func ExampleWithFunction() {
	query, err := jq.Parse(".[] | f | f(3)")
	if err != nil {
		log.Fatalln(err)
	}
	code, err := jq.Compile(
		query,
		jq.WithFunction("f", 0, 1, func(x interface{}, xs []interface{}) interface{} {
			if x, ok := toFloat(x); ok {
				if len(xs) == 1 {
					if y, ok := toFloat(xs[0]); ok {
						x *= y
					} else {
						return fmt.Errorf("f cannot be applied to: %v, %v", x, xs)
					}
				} else {
					x += 2
				}
				return x
			}
			return fmt.Errorf("f cannot be applied to: %v, %v", x, xs)
		}),
	)
	if err != nil {
		log.Fatalln(err)
	}
	input := []interface{}{0, 1, 2.5, json.Number("10000000000000000000000000000000000000000")}
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
	// 6
	// 9
	// 13.5
	// 3e+40
}
