.PHONY: infra-up infra-down migrate

infra-up:
	docker-compose -f docker-compose-infra.yml up -d --wait

infra-down:
	docker-compose -f docker-compose-infra.yml down

migrate:
	cd backend && go run ./cmd/migrate/

tidy:
	cd backend && go mod tidy

fmt:
	cd backend && go fmt ./...
	