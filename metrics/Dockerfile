FROM python:3.10-slim

WORKDIR /app

COPY exporter.py .

RUN pip install --no-cache-dir prometheus_client kubernetes

CMD ["python", "exporter.py"]
