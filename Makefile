.PHONY: run
run:
	templ generate
	go run cmd/templ-demo/main.go