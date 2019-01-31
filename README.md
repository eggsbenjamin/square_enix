## Square Enix Tech Test

### Dependencies

- docker
- docker-compose
- make
- dep

### Setup

Install Deps: `dep ensure`

Run tests : `make docker_test`

Start server

```
MYSQL_USER=root \
MYSQL_HOST=localhost \
MYSQL_DB=square_enix \
BATCH_SIZE=20 \
POLL_INTERVAL=5 \
PORT=8080 \
go run ./cmd/*
```

All the env vars should be self-explanatory except for `POLL_INTERVAL` which is the time in seconds between the polling of the Process table for work.

### Design

See [design notes](./assets/notes.pdf).

### Data

To add elements to be processed, add entries to the `Element` table.

`INSERT INTO Element (data) VALUES ('test');`

### TODO

- creates schema [x]
- enable tests to be run inside container [x]
- create and test repositories [x]
- create and test unit of work [x]
- create and test http handlers []
- create e2e tests []

