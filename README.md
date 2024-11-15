# Meilisearch

Script used at [SneakersAPI.dev](https://sneakersapi.dev) to index Meilisearch from PostgreSQL.

**Key features:**

- Replicates data from PostgreSQL to Meilisearch.
- Should support almost any types, the inserted data is in JSON.
- Time-series data can be synced via cursor to avoid full table scans.

**About performance:**

- 180k rows with complex nested JSON: 45s, about ~3.9k rows/s (async enabled)

## Configuration

Configuration is done via a YAML file. See `config.example.yml` for reference.

**About `enable_async` and `wait_time`:**

- When `enable_async` is true, batches are sent in a separate thread, it greatly improves performance but the memory usage of the MeiliSearch instance is much higher.
- When `enable_async` is false, `wait_time` is the time in milliseconds to wait before sending the next batch. This is useful on low-end VPS to avoid memory issues as the MeiliSearch instance will have time to free up memory between batches.
- For low-memory VPS, it's recommended to use `enable_async: false` and to set a `wait_time`, depending on the number of rows to sync. You might be interested to use these environment variables on the MeiliSearch instance also:
  - `MEILI_EXPERIMENTAL_REDUCE_INDEXING_MEMORY_USAGE=true`
  - `MEILI_EXPERIMENTAL_MAX_NUMBER_OF_BATCHED_TASKS=5`

## Running

```bash
export MEILISEARCH_DSN=<meilisearch_dsn>
export MEILISEARCH_API_KEY=<meilisearch_api_key>
export DATABASE_URL=<database_url>

go run . [-only=<index_name>] [-drop=[true|false]] [-meta=[true|false]] [-config=<path>]
```

- `-only=<index_name>`: Avoid running all indexes and only process the one specified, by its destination name.
- `-drop=[true|false]`: Drop the index and reset cursor, if any.
- `-meta=[true|false]`: Update index metadata (searchable, filterable, sortable).
- `-config=<path>`: Path to the configuration file. Defaults to `config.yml`.

## Docker

```bash
docker build -t meilisearch .
docker run -e MEILISEARCH_DSN=<meilisearch_dsn> \
    -e MEILISEARCH_API_KEY=<meilisearch_api_key> \
    -e DATABASE_URL=<database_url> \
    meilisearch \
    [-only=<index_name>] \
    [-drop=[true|false]] \
    [-meta=[true|false]] \
    [-config=<path>]
```
