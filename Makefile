.PHONY: run
run:
	templ generate
	go run cmd/templ-demo/main.go

fmt:
	templ fmt ./internal/app/views
	go fmt ./...