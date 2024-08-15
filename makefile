SHELL := /bin/bash
.DEFAULT_GOAL := run
DOMAIN := http://localhost:8080
PROD_DOMAIN := https://gt5qrserver-kwtxn.ondigitalocean.app

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

setup-prod:
	curl "$(PROD_DOMAIN)/add-team?team=Alpacas"
	curl "$(PROD_DOMAIN)/add-team?team=Bisons"
	curl "$(PROD_DOMAIN)/add-team?team=Coyotes"
	curl "$(PROD_DOMAIN)/add-team?team=Dolphins"
	curl "$(PROD_DOMAIN)/add-team?team=Eagles"

list-prod-teams:
	curl "$(PROD_DOMAIN)/list-teams"