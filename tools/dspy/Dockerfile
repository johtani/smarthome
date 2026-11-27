FROM golang:1.26 AS catalog-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY internal ./internal
COPY subcommand ./subcommand
COPY server/cron/influxdb ./server/cron/influxdb
COPY tools/resolver-catalog ./tools/resolver-catalog
RUN CGO_ENABLED=0 go build -o /out/resolver-catalog ./tools/resolver-catalog

FROM python:3.11-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /opt/dspy
COPY --from=catalog-build /out/resolver-catalog /opt/dspy/resolver-catalog

COPY tools/dspy/requirements.txt /opt/dspy/requirements.txt
RUN pip install --no-cache-dir -r /opt/dspy/requirements.txt

COPY tools/dspy/prepare_dataset.py /opt/dspy/prepare_dataset.py
COPY tools/dspy/optimize_and_evaluate.py /opt/dspy/optimize_and_evaluate.py
COPY tools/dspy/command_catalog.sample.json /opt/dspy/command_catalog.sample.json
COPY tools/dspy/run-batch.sh /opt/dspy/run-batch.sh
COPY tools/dspy_common /opt/dspy_common

RUN chmod +x /opt/dspy/run-batch.sh

ENTRYPOINT ["/opt/dspy/run-batch.sh"]
