module github.com/providenceinnovation/c20-alexa-skills

go 1.16

replace github.com/tbolsh/extend-go-nginx-postgres-docker/persistense v0.0.0 => ./src/persistense

require (
	github.com/lib/pq v1.10.5 // indirect
	github.com/tbolsh/extend-go-nginx-postgres-docker/persistense v0.0.0
)
