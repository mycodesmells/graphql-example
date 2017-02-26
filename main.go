package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/graphql-go/graphql"
)

type Input struct {
	Key int `json:"key"`
}

var inputType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Input",
	Fields: graphql.Fields{
		"key": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return 42, nil
			},
		},
	},
})

func main() {
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				fmt.Println(p.Args)
				return "world", nil
			},
		},
		"answer": &graphql.Field{
			Type: inputType,
			// Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// 	key := 42 * p.Args["mul"].(int)
			// 	return Input{Key: key}, nil
			// },
			// Args: graphql.FieldConfigArgument{
			// 	"mul": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
			// },
		},
	}

	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	// var query string
	// fmt.Scanf("%s", &query)
	// query := "{answer(mul: 5){key}}"
	query := "{hello}"
	// query := "{answer(id: \"aaa\")}"

	params := graphql.Params{Schema: schema, RequestString: query}
	r := graphql.Do(params)
	if r.HasErrors() {
		log.Fatalf("Failed due to errors: %v\n", r.Errors)
	}

	rJSON, err := json.Marshal(r)
	log.Printf("Found: %s", rJSON)
}
