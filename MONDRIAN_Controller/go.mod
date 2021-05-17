module controller/api_test

go 1.14

replace controller/handler => ./handler

replace controller/db => ./db

replace controller/types => ./types

replace controller/api => ./api

replace controller/config => ./config

require (
	controller/config v0.0.0-00010101000000-000000000000
	controller/db v0.0.0-00010101000000-000000000000
	controller/handler v0.0.0-00010101000000-000000000000
	controller/types v0.0.0-00010101000000-000000000000
	github.com/mattn/go-sqlite3 v1.14.7
)
