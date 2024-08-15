SHELL := /bin/bash
.DEFAULT_GOAL := run
DOMAIN := http://localhost:8080

run:
	source default.env && go run main.go

fill-teams:
	curl "$(DOMAIN)/add-team?team=alpa"
	curl "$(DOMAIN)/add-team?team=beta"
	curl "$(DOMAIN)/add-team?team=gamma"

list-teams:
	curl "$(DOMAIN)/list-teams"

access-qr-code:
	echo "$(DOMAIN)/4nj92jh"