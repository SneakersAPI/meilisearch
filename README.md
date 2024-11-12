# Meilisearch

Script used at [SneakersAPI.dev](https://sneakersapi.dev) to index Meilisearch from PostgreSQL.

**Key features:**

- Replicates data from PostgreSQL to Meilisearch.
- Should support almost any types, the inserted data is in JSON.
- Time-series data can be synced via cursor to avoid full table scans.

**About performance:**

- 180k rows with complex nested JSON: 45s, about ~3.9k rows/s

## Configuration

Configuration is done via a YAML file. See `config.example.yml` for reference.

## Running

```bash
go run . [-only=<index_name>] [-drop=[true|false]] [-meta=[true|false]] [-config=<path>]
```

- `-only=<index_name>`: Avoid running all indexes and only process the one specified, by its destination name.
- `-drop=[true|false]`: Drop the index and reset cursor, if any.
- `-meta=[true|false]`: Update index metadata (searchable, filterable, sortable).
- `-config=<path>`: Path to the configuration file. Defaults to `config.yml`.

## Docker

```bash
docker build -t meilisearch .
docker run meilisearch [-only=<index_name>] [-drop=[true|false]] [-meta=[true|false]] [-config=<path>]
```
