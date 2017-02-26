# Advanced GraphQL in Golang

In the previous post, we've seen how easy it is to set up a simple GraphQL server in Golang. The problem is, that we have hardly done anything special, that would distinguish our API from one written for REST. It's time for something more advanced, that is nested objects and fetching them from different sources.

# Defining complex types

Our first field in the GraphQL API looked very simple:

    fields := graphql.Fields{
        "hello": &graphql.Field{
            Type: graphql.String,
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                return "world", nil
            },
        },
    }

Returning primitive (_scalar_ in the GraphQL world) values is definitely not enough for any real use purposes. Fortunately, we can create more complicated structures. Imagine we have a user that has three attributes: login (string), admin (boolean) and active (boolean). In order to expose such entity to the API users, we need to create an _object_ and save it to a type:

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
            }
        },
    })

As you can see, we give our newly created type a name, define all the fields and their types. Now, we can add such field to the API:

    fields := graphql.Fields{
        ...
        "user": &graphql.Field{
            Type: userType,
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                ...
            },
        },
    }

Right, but how do we resolve an object of our type? With Go it's just as simple as returning an instance of a struct, that has all the necessary fields! Let's define the struct first:

    type User struct {
        Login  string `json:"login"`
        Admin  string `json:"admin"`
        Active string `json:"active"`
    }

Let's pretend that in our real example we keep our users data in Postgres database. For simplicity (spoiler alert, we'll alter it in the future), we'll ask for all users but use only the first one, as there will be just the one. In this case, our `Resolve` function could look like this:

    ...
    Resolve: func(p graphql.ResolveParams) (interface{}, error) {
        rows, qerr := db.Query("SELECT * from users")
        if qerr != nil {
            log.Fatalf("Failed to read from the database: %v", err)
        }
        var u User
        rows.Next()
        err = rows.Scan(&u.Login, &u.Admin, &u.Active)
        if err != nil {
            log.Fatalf("Failed to load user data: %v", err)
        }

        return u, nil
    },
    ...

Now if we run the application, we can ask the API for the user's login:

    $ go run main.go -query="{user{login}}"
    2017/02/23 23:36:21 Found: {"data":{"user":{"login":"slomek"}}}

We can, obviously, ask for other fields if that's what we need:

    $ go run main.go -query="{user{login,admin,active}}"
    2017/02/23 23:37:02 Found: {"data":{"user":{"active":"true","admin":"true","login":"slomek"}}}

# Inserting a relation

This was fun, but still, we haven't done anything special. That's why we need to enhance our example a bit, and add another data source that would extend the information about our users. For example, we could store user's profile in another database, eg. in MongoDB. In our case, it will keep the users' login names and their permissions. How could we use this data in our API?

We might have a preference to hide the presence of another database from our users and return `permissions` as just another field in the response. First, we need to add another subfield to `user` field:

    var userType = graphql.NewObject(graphql.ObjectConfig{
        Name: "User",
        Fields: graphql.Fields{
            ...
            "permissions": &graphql.Field{
                Type: graphql.NewList(graphql.String),
            },
        },
    })

As you can see, we gave it a type of a list of strings, as this is what we have in the database. However, we don't want to connect to Mongo at `user`'s level and in its `Resolve` function. Fortunately, we can the resolve part inside subfield, just by adding `Resolve` function of its own:

    ...
    "permissions": &graphql.Field{
        Type: graphql.NewList(graphql.String),
        Resolve: func(p graphql.ResolveParams) (interface{}, error) {
            var profile UserProfile
            mdb.C("profiles").Find(bson.M{}).One(&profile)
            return profile.Permissions, nil
        },
    },
    ...

Now we can query for permissions as well, but now we'd be getting data from multiple sources at once:

    $ go run main.go -query="{user{login,permissions}}"
    2017/02/23 23:46:04 Found: {"data":{"user":{"login":"slomek","permissions":["divide","conquer"]}}}

# Working example

The best part of having multiple sources divided so easily in the GraphQL is that we don't query MongoDB if we don't have to! We could do that in REST as well, but that would probably be much more complicated, but here is as simple as it can get. Running query for Postgres data results in MongoDB left alone:

    $ go run main.go -query="{user{login}}"

    # MONGO CONSOLE:
    mongo_1       | 2017-02-23T22:49:49.319+0000 I NETWORK  [thread1] connection accepted from 172.20.0.1:44826 #26 (2 connections now open)
    mongo_1       | 2017-02-23T22:49:49.319+0000 I COMMAND  [conn26] command admin.$cmd command: getnonce { getnonce: 1 } numYields:0 reslen:65 locks:{} protocol:op_query 0ms
    mongo_1       | 2017-02-23T22:49:49.319+0000 I COMMAND  [conn26] command admin.$cmd command: isMaster { ismaster: 1 } numYields:0 reslen:189 locks:{} protocol:op_query 0ms
    mongo_1       | 2017-02-23T22:49:49.320+0000 I COMMAND  [conn26] command admin.$cmd command: ping { ping: 1 } numYields:0 reslen:37 locks:{} protocol:op_query 0ms
    mongo_1       | 2017-02-23T22:49:49.329+0000 I -        [conn26] end connection 172.20.0.1:44826 (2 connections now open)

If we ask for permission data, we do see some traffic there:

    $ go run main.go -query="{user{login,permissions}}"
    # MONGO CONSOLE:
    mongo_1       | 2017-02-26T22:51:09.355+0000 I NETWORK  [thread1] connection accepted from 172.20.0.1:44830 #27 (2 connections now open)
    mongo_1       | 2017-02-26T22:51:09.355+0000 I COMMAND  [conn27] command admin.$cmd command: getnonce { getnonce: 1 } numYields:0 reslen:65 locks:{} protocol:op_query 0ms
    mongo_1       | 2017-02-26T22:51:09.356+0000 I COMMAND  [conn27] command admin.$cmd command: isMaster { ismaster: 1 } numYields:0 reslen:189 locks:{} protocol:op_query 0ms
    mongo_1       | 2017-02-26T22:51:09.356+0000 I COMMAND  [conn27] command admin.$cmd command: ping { ping: 1 } numYields:0 reslen:37 locks:{} protocol:op_query 0ms
    mongo_1       | 2017-02-26T22:51:09.362+0000 I COMMAND  [conn27] command graphql-example.profiles command: find { find: "profiles", filter: {}, skip: 0, limit: 1, batchSize: 1, singleBatch: true } planSummary: COLLSCAN keysExamined:0 docsExamined:1 cursorExhausted:1 numYields:0 nreturned:1 reslen:183 locks:{ Global: { acquireCount: { r: 2 } }, Database: { acquireCount: { r: 1 } }, Collection: { acquireCount: { r: 1 } } } protocol:op_query 0ms
    mongo_1       | 2017-02-26T22:51:09.366+0000 I -        [conn27] end connection 172.20.0.1:44830 (2 connections now open)

The full source code of this example is available [on Github](https://github.com/mycodesmells/graphql-example).
