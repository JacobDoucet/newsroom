.PHONY: dev down clean logs

dev:
	docker compose up --build

down:
	docker compose down

clean:
	docker compose down -v

logs:
	docker compose logs -f
