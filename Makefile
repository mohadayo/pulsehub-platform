.PHONY: test test-go test-python test-ts up down build lint

test: test-go test-python test-ts
	@echo "All tests passed."

test-go:
	cd services/collector && go test -v ./...

test-python:
	cd services/analytics && pip install -q -r requirements.txt && pytest -v

test-ts:
	cd services/gateway && npm install --silent && npm test

lint: lint-python lint-ts
	@echo "All lints passed."

lint-python:
	cd services/analytics && pip install -q flake8 && flake8 --max-line-length=120 app.py

lint-ts:
	cd services/gateway && npx tsc --noEmit

up:
	docker compose up --build -d

down:
	docker compose down

build:
	docker compose build

logs:
	docker compose logs -f
