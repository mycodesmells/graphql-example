package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/graphql-go/graphql"
	_ "github.com/lib/pq"
)

// User .
type User struct {
	Login  string `json:"login"`
	Admin  string `json:"admin"`
	Active string `json:"active"`
}

type UserProfile struct {
	Permissions []string `json:"permissions"`
}

type userDetails struct {
	username string
	admin    bool
	active   bool
}

var userType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.Fields{
		"login": &graphql.Field{
			Type: graphql.String,
		},
		"admin": &graphql.Field{
			Type: graphql.String,
		},
		"active": &graphql.Field{
			Type: graphql.String,
		},
		"permissions": &graphql.Field{
			Type: graphql.NewList(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var profile UserProfile
				mdb.C("profiles").Find(bson.M{}).One(&profile)
				return profile.Permissions, nil
			},
		},
	},
})

var query string
var mdb *mgo.Database

func init() {
	flag.StringVar(&query, "query", "{}", "query to ask server for data")
}

func main() {
	flag.Parse()

	db, err := sql.Open("postgres", "user=postgres dbname=graphql sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	mdb = session.DB("graphql-example")

	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				fmt.Println(p.Args)
				return "world", nil
			},
		},
		"user": &graphql.Field{
			Type: userType,
			Args: graphql.FieldConfigArgument{
				"login": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				login := p.Args["login"].(string)

				rows, qerr := db.Query("SELECT * from users WHERE username = $1", login)
				if qerr != nil {
					log.Fatalf("Failed to read from the database: %v", err)
				}
				var u User
				exist := rows.Next()
				if exist {
					err = rows.Scan(&u.Login, &u.Admin, &u.Active)
					if err != nil {
						log.Fatalf("Failed to load user data: %v", err)
					}
				}

				return u, nil
			},
		},
	}

	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}

	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(rootQuery),
	}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	params := graphql.Params{Schema: schema, RequestString: query}
	r := graphql.Do(params)
	if r.HasErrors() {
		log.Fatalf("Failed due to errors: %v\n", r.Errors)
	}

	rJSON, _ := json.Marshal(r)
	log.Printf("Found: %s", rJSON)
}
