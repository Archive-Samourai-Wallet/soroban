FROM python:3.7.5-alpine3.10
RUN apk --no-cache --update add ca-certificates
RUN apk add --no-cache alpine-sdk libffi-dev openssl-dev

ENV PYTHONUNBUFFERED=1 PYTHONHASHSEED=random PYTHONDONTWRITEBYTECODE=1

WORKDIR /app
COPY requirements.txt /app
WORKDIR /app
RUN pip install -r requirements.txt

COPY . /app

ENTRYPOINT ["python", "soroban-client.py"]
