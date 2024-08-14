SHELL := /bin/bash
.DEFAULT_GOAL := run

run:
	source default.env && go run main.go