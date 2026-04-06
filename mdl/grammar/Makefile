# Makefile for MDL grammar generation
#
# Usage:
#   make generate    - Generate Go parser from MDLLexer.g4 and MDLParser.g4
#   make clean       - Remove generated files
#
# Prerequisites:
#   - ANTLR4 must be installed (https://www.antlr.org/)
#   - macOS:  brew install antlr4
#   - Linux:  apt install antlr4  (or download JAR manually)
#   - Or use: pip install antlr4-tools

LEXER = MDLLexer.g4
PARSER = MDLParser.g4
OUTPUT_DIR = parser
PACKAGE = parser

# Find antlr4 command (try common locations)
ANTLR4 := $(shell which antlr4 2>/dev/null || which antlr 2>/dev/null)

.PHONY: generate clean check-antlr

check-antlr:
ifndef ANTLR4
	$(error ANTLR4 not found. Install with: brew install antlr4 (macOS) or pip install antlr4-tools)
endif

generate: check-antlr $(LEXER) $(PARSER)
	@mkdir -p $(OUTPUT_DIR)
	$(ANTLR4) -Dlanguage=Go -no-visitor -package $(PACKAGE) -o $(OUTPUT_DIR) $(LEXER) $(PARSER)
	@echo "Generated Go parser in $(OUTPUT_DIR)/"

clean:
	rm -rf $(OUTPUT_DIR)
