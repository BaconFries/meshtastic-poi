# Performance Notes

## Goals

| Operation | Target (100k points) |
|-----------|---------------------|
| Validation | < 2 seconds |
| Optimization | < 5 seconds |
| Radius search | < 100 ms |

## Strategies

### Spatial Indexing

The quadtree index (`internal/spatial/index.go`) pre-indexes feature centroids for O(log n) candidate lookup before precise Haversine distance checks.

### Memory

- Feature collections are processed in-place where safe (normalize, property cleanup)
- Pipeline stages return the same slice backing array when possible
- No CGO avoids cross-boundary allocation overhead

### ArcGIS Pagination

Pages are fetched sequentially with `resultOffset` / `resultRecordCount` sized to `maxRecordCount`. Features are appended to a single slice with pre-allocated capacity.

### Deduplication

Coordinate dedup uses string keys at 6 decimal precision (~0.1m). ID dedup uses a hash map.

### Benchmarking

Run benchmarks as the project grows:

```bash
go test -bench=. -benchmem ./internal/...
```

Consider adding large synthetic datasets under `testdata/` for regression testing.

## Optimization Tips

- Use `--minimal` for smallest output size
- Use `--dedupe` to remove redundant points before export
- Split large datasets with `--max-features` for device constraints
