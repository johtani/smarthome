FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /opt/dspy-resolver

COPY tools/dspy-resolver/requirements.txt /opt/dspy-resolver/requirements.txt
RUN pip install --no-cache-dir -r /opt/dspy-resolver/requirements.txt

COPY tools/dspy-resolver/app.py /opt/dspy-resolver/app.py

EXPOSE 8080

CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8080"]
