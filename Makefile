
dev:
	docker-compose up --build

test:
	docker-compose -f docker-compose.test.yaml up --build

bench:
	go run cmd/benchmark/main.go

rebalance:
	go run cmd/rebalance/main.go
