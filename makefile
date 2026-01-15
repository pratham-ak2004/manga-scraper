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