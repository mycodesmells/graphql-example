# Making Changes in GraphQL API

So far we've seen how to build a GraphQL API that serves the user with data stored in the databases. But what about making changes? With a few tweaks in the structure, we can achieve this as well. Let's take a look on things called mutations.

# Mutations

In GraphQL, any request sent to the server with an intention of making some modifications on the underlying data is called _a mutation_. While it is structured the same way as any other _query_, it needs to be explicitly marked as mutation not to cause some accidental changes.

As of now, our server has a tiny schema, consisting just of fields that can be returned by queries:

    schemaConfig := graphql.SchemaConfig{
        Query: graphql.NewObject(rootQuery),
    }

In the [previous post](http://mycodesmells.com/post/being-selective-with-graphql) we made it possible to query for a specific user by their login name. Now, we'd like to add a permission for that user.

We start by adding a root mutation to our schema:

        rootMutation := graphql.ObjectConfig{Name: "RootMutation", Fields: mutations}

    schemaConfig := graphql.SchemaConfig{
        Query:    graphql.NewObject(rootQuery),
        Mutation: graphql.NewObject(rootMutation),
    }

# Mutation vs Query

Actually, from now on we can build the structure as if it was a plain query. An `ObjectConfig` we create is, after all, the same type we used for building `rootQuery` object. It is our responsibility now, as API creators, to provide appropriate functionality to the fields.

# Building it

With mutations, everything that needs to be changed happens in `Resolve` function, but we still need to define a type of the field. This way we can give our users some feedback about the state of their change. We can return an user object with updated fields, but as we are changing just a part of the user, let's return just the flag saying whether the change was successful or not:

    mutations := graphql.Fields{
        "addPermission": &graphql.Field{
            Type: graphql.Boolean,
            Args: graphql.FieldConfigArgument{
                "login": &graphql.ArgumentConfig{
                    Type: graphql.NewNonNull(graphql.String),
                },
                "permission": &graphql.ArgumentConfig{
                    Type: graphql.NewNonNull(graphql.String),
                },
            },
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                login := p.Args["login"].(string)
                permission := p.Args["permission"].(string)
    
                err := mdb.C("profiles").Update(bson.M{"_id": login}, bson.M{
                    "$addToSet": bson.M{
                        "permissions": permission,
                    },
                })
    
                return err == nil, err
            },
        },
    }
    
# Seeing it work

First, let's see how our user profile looks:

    $ go run main.go -query "{user(login: \"slomek\"){ login, admin, active, permissions }}"
    2017/03/11 13:57:51 Found: {"data":{"user":{"active":"true","admin":"true","login":"slomek","permissions":["divide","conquer"]}}}

Now let's make a change and add _doing cool demos_ permission:

    $ go run main.go -query "mutation { addPermission(login: \"slomek\", permission: \"doing cool demos\") }"
    2017/03/11 13:58:17 Found: {"data":{"addPermission":true}}
    
Since it worked, we shall see an updated profile with another query:

    $ go run main.go -query "{user(login: \"slomek\"){ login, admin, active, permissions }}"
    2017/03/11 13:58:21 Found: {"data":{"user":{"active":"true","admin":"true","login":"slomek","permissions":["divide","conquer","doing cool demos"]}}}

Hurray!
