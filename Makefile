DB_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable

postgres:
	docker run --name postgres --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:14-alpine

createdb:
	docker exec -it postgres createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres dropdb simple_bank

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migrateup_n:
	@read -p "Enter number of migrations: " n; \
	migrate -path db/migration -database "$(DB_URL)" -verbose up $$n

migratedown_n:
	@read -p "Enter number of migrations: " n; \
	migrate -path db/migration -database "$(DB_URL)" -verbose down $$n

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mockery:
	mockery --name=Store --recursive

proto: 
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
	proto/*.proto

evans:
	evans --host localhost --port 9000 -r repl

redis:
	docker run --name redis --network bank-network -p 6379:6379 -d redis:7-alpine

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mockery migrateup_n migratedown_n proto evans redis
