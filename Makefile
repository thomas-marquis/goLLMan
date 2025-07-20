help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  generate    Generate templates using 'templ generate'."
	@echo "  ui          Start the UI Genkit  server."
	@echo "  documents/%.epub      Uncompress an epub file located in the documents folder. "
	@echo "  					   Use the full file path without extension e.g.: 'documents/my_file'"
	@echo "  help       Show this help message."
	@echo ""

generate:
	@templ generate
.PHONY: generate

ui:
	@genkit start -- go run main.go chat -i http -v
.PHONY: ui

documents/%.zip:
	@echo "Creating a zip copy of $*.epub"
	@cp documents/$*.epub $@

documents/%: documents/%.zip
	@echo "Extracting $@.epub into documents/$*"
	@unzip $< -d documents/$*
	@rm $<