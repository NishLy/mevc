include .env
export

.PHONY: run migrate-diff

run:
	air 

migrate-diff:
	go run ./cmd/atlas/diff $(name)

migrate-hash:
	atlas migrate hash

migrate-apply:
	go run ./cmd/atlas/apply