start:
	python app.py

worker:
	python app.py --worker

listener:
	python app.py --listener

flower:
	python app.py --flower

compose:
	@source .env.docker && docker compose up -d

dump-db:
	DB_CONTAINER=manga-scraper-postgres
	@source .env.docker && docker exec -t $${DB_CONTAINER} pg_dump -U $${DB_USER} $${DB_NAME} > dump.sql