# GraphQL example in Go

Start databases:

    $ docker-compose up

Run API:

    $ go run main.go -query=YOUR_QUERY_HERE

Example:

    $ go run main.go -query={hello}
    2017/02/23 23:43:20 Found: {"data":{"hello":"world"}}
