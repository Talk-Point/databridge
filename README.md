# DataBridge

ETL to from various data sources to a data warehouse. Start to build the own etl tool, why no other tool we found bring the performance or has a connector to time series databases.

## Usage

Best case is to run this with docker image

```sh
docker run --rm \
  -e API_TOKEN="<api-key>" \
  -e TIMESCALEDB_CONN_STR="postgresql://postgres:password@localhost:5432/postgres" \
  talkpoint/databridge \
  -config "examples/sage_khk_vk_belege.yaml" \
  -run-schema \
  -log-level debug \
  -date 2024-09-25
```

```sh
$ go run cmd/databridge/main.go -config "runs/sage_khk_vk_beleg.yaml" -start "2024-09-25T00:00:00Z" -end "2024-09-25T23:59:59Z"
$ go run cmd/databridge/main.go -config "runs/sage_khk_vk_beleg.yaml" -interval 30m
```

## Params

- `-config` Path to the configuration file
- partition interval
    - `-interval` Interval in minutes
    - `-start` Start time in RFC3339 format
    - `-end` End time in RFC3339 format
    - `-date` Date like 2024-09-25
- `-run-schema` Run schema
- `-dry-run` Dry run mode
- `-log-level` Log level (default "info")