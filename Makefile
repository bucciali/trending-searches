run:
	go run cmd/main.go
down:
	docker compose down
up:
	docker-compose up -d
dup:
	docker-compose up -d --build