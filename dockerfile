FROM python:3.14.2

WORKDIR /app

COPY requirements.txt .
COPY .env.docker .env

RUN apt-get update && apt-get install -y make curl unzip xvfb
RUN pip install --no-cache-dir -r requirements.txt
RUN playwright install --with-deps chromium

COPY . .

CMD xvfb-run -a -s "-screen 0 1920x1080x24" python app.py --worker