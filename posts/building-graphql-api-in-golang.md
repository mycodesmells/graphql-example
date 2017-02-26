# Building GraphQL API in Go

For a long time building an API (eg. for web applications) has become associated with REST. Meanwhile, less than two years ago, Facebook announced a technology they've been using for some time called GraphQL. It's an alternative way to provide data to the users, where each client can squeeze multiple queries into a single one.

# What is GraphQL

At first, looking at GraphQL might be a bit overwhelming for the newcomers. In REST, the responses may be in a form of JSONs (if you choose so), but clients' queries take many forms, as it can be composed of HTTP method, path parameters and body content. The idea of request-response stays unchanged in GraphQL, but the way the communication flows is slightly different. It all starts with the query, which in fact is a template that needs to be filled with data by the server. For example, if we want to get details of _current user_, we might build the following query:

    currentUser {
        username
        role
        lastLogin
    }

As you can see, it looks like an empty JSON, where keys are defined, but values are not provided. It is also simplified, as you don't need commas to separate field names (as long as they are not in the same line). Writing it in a single line needs a few modifications:

    currentUser { username, role, permissions }

The important thing to note is that if _currentUser_ response is an object, you need to define what fields are required to be returned. Our server then can response with a filled template:

    currentUser {
        username: "someuser"
        role: "admin"
        lastLogin: "2017-02-19 12:13:14"
    }

GraphQL is often used as another level of abstraction between a client and the data sources, especially if data is fetched from more than one place. For example, you can have one database storing current user's roles in one database, while the history of user's logins in another.

# Simple example

In order to build a simple GraphQL API, we need two main elements: the definition of API structure, and the way to fill query template with some data. First, we define what the users can ask for using `graphql.Fields` and provide a specific type for each part. In our case, we'll expose a single field called _hello_, to which the server respond with a string, "`world`":

    fields := graphql.Fields{
        "hello": &graphql.Field{
            Type: graphql.String,
            ...
        }
    }

In order to make our server respond the way we want, we need to provide a _resolve_ function for the field, which will fill query template with our response:

    ...
    "hello": &graphql.Field{
        Type: graphql.String,
        Resolve: func(p graphql.ResolveParams) (interface{}, error) {
            return "world", nil
        },
    }
    ...

As you can see, it's as simple as it can get. In order to run the GraphQL server we need to build a schema using so called _queries_ which need to have a name (eg. in our case _RootQuery_) and fields definitions (created a moment ago):

    rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
    schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
    schema, err := graphql.NewSchema(schemaConfig)

Finally, we can query the server with simple string argument:

    params := graphql.Params{Schema: schema, RequestString: "{hello}"}
    r := graphql.Do(params)
    if r.HasErrors() {
        log.Fatalf("Failed due to errors: %v\n", r.Errors)
    }

    rJSON, err := json.Marshal(r)
    log.Printf("Found: %s", rJSON)

What we get in response is a correct JSON object with all information stored under `data` field:

    $ go run main.go
    2017/02/19 20:10:51 Found: {"data":{"hello":"world"}}

The next steps would be to create more sophisticated data structures, add several data sources and expose it via some web application, to match the features of REST.
