# Being Selective With GraphQL

Building an API in GraphQL looks very easy, as you could have already seen based on two previous posts. The thing is that everything looks simple when no parameters are provided, right? Let's take a look at how to make our API just a bit more dynamic.

In the two previous posts we've seen how to create [a simple API](http://mycodesmells.com/post/building-graphql-api-in-go) and the one with [slightly more complicated](http://mycodesmells.com/post/advanced-graphql-in-golang) item types, but we've never seen any parameters provided to the query. What if we want to fetch data about one specific _user_?

### Query Parameters

Adding parameters to the query in GraphQL is in fact still very simple. First of all, we need to define all arguments that can be applied to given branch of our API. In our case, we'd like to provide _login_ value when asking for the user, to make sure we return correct one. In order to do that we just add another property on `graphql.Fields` item:

    fields := graphql.Fields{
        ...
        "user": &graphql.Field{
            Type: userType,
            Args: graphql.FieldConfigArgument{
                "login": &graphql.ArgumentConfig{
                    Type: graphql.NewNonNull(graphql.String),
                },
            },
            ...
        },
    }
    
We need to define the `Type` for each argument, and we can wrap it in `graphql.NewNonNull(..)` to make it required.

Next step is to alter the `Resolve` function for this field, to take our newly added parameter into consideration:

    "user": &graphql.Field{
        ...
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

And that's it! Our GraphQL API now requires you to add an argument when querying for the users. The last thing that needs to be changed is the way we query for data. Previously we could send a query like:

    {
        user {
            login
            admin
            active
        }
    }
    
Now our users need to treat `user` as a function to which we provide named parameters:

    {
        user(login: "slomek") {
            login
            admin
            active
        }
    }

Our response looks exactly as it was before, but now we can have more users in the DB and ask for a specific one:

    $ go run main.go -query "{user(login: \"slomek\"){login, admin, active}}"
    2017/03/05 20:51:43 Found: {"data":{"user":{"active":"true","admin":"true","login":"slomek"}}}
    
The source code of this example is available [on Github](https://github.com/mycodesmells/graphql-example).
