include .env
export

migrate-create:
	sudo docker run --rm \
	-v $(shell pwd)/migrations:/migrations \
	migrate/migrate \
	create \
	-ext sql \
	-dir /migrations \
	-seq $(name)
	sudo chown -R $(shell id -u):$(shell id -g) $(shell pwd)/migrations

migrate-up:
	docker run --rm \
	--network result-lookup_default \
	-v $(shell pwd)/migrations:/migrations \
	migrate/migrate \
	-source file://migrations \
	-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
	-verbose \
	up

migrate-down:
	docker run --rm \
	--network result-lookup_default \
	-v $(shell pwd)/migrations:/migrations \
	migrate/migrate \
	-source file://migrations \
	-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
	-verbose \
	down 1

up:
	docker compose up -d --build postgres redis api

down:
	docker compose down

precompute:
	docker compose run --rm precompute --year=$(year)

db:
	docker exec -it $(POSTGRES_HOST) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

redis:
	docker exec -it redis redis-cli -a $(REDIS_PASSWORD)

logs:
	docker logs -f result-lookup-api

clean:
	docker compose down --remove-orphans -v

clean-hard:
	docker compose down --remove-orphans -v --rmi "all"