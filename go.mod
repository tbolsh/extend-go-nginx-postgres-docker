module github.com/providenceinnovation/c20-alexa-skills

go 1.16

replace (
	github.com/tbolsh/extend-go-nginx-postgres-docker/genericjson v0.0.0 => ./src/genericjson
	github.com/tbolsh/extend-go-nginx-postgres-docker/persistense v0.0.0 => ./src/persistense
)

require (
	github.com/gorilla/mux v1.8.0
	github.com/lib/pq v1.10.5 // indirect
	github.com/tbolsh/extend-go-nginx-postgres-docker/genericjson v0.0.0
	github.com/tbolsh/extend-go-nginx-postgres-docker/persistense v0.0.0
)
