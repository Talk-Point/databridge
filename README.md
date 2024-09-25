# DataBridge

ETL to from various data sources to a data warehouse.

## Usage

```sh
$ go run cmd/databridge/main.go -config "runs/sage_khk_vk_beleg.yaml" -start "2024-09-23T00:00:00Z" -end "2024-09-23T00:00:00Z"
$ go run cmd/databridge/main.go -config "runs/sage_khk_vk_beleg.yaml" -interval 30m
```

## Params

- `-config` Path to the configuration file
- partition interval
    - `-interval` Interval in minutes
    - `-start` Start time in RFC3339 format
    - `-end` End time in RFC3339 format
- `-run-schema` Run schema
- `-dry-run` Dry run mode
- `-log-level` Log level (default "info")