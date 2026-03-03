.PHONY: infra-up infra-down migrate tidy fmt test run-backend install-frontend run-frontend

infra-up:
	docker-compose -f docker-compose-infra.yml up -d --wait

infra-down:
	docker-compose -f docker-compose-infra.yml down --volumes --remove-orphans

migrate:
	cd backend && go run ./cmd/migrate/

tidy:
	cd backend && go mod tidy

fmt:
	cd backend && go fmt ./...

test:
	cd backend && go test -v ./...

run-backend:
	cd backend && go run ./cmd/api

install-frontend:
	cd frontend && npm install

run-frontend: install-frontend 
	cd frontend && npm run dev