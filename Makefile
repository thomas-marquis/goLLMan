generate:
	@templ generate

ui:
	@genkit start -- go run main.go chat -i http -v

.PHONY: generate ui