# Performance Tuning for Small Servers (1 vCPU / 2 GB RAM)

## Problem: PostgreSQL CPU spikes to 100% after several hours

This is typically caused by:
1. **Too many database connections** overwhelming the single CPU core
2. **Missing indexes** on the `logs` table causing full table scans
3. **Unoptimized aggregate queries** that run repeatedly (dashboard stats)
4. **Autovacuum lag** causing table bloat and slower queries over time

## Problem: High CPU usage during high request rate (rpm)

When request volume is high, CPU spikes are usually from:
1. **Token counting overhead** - tiktoken encoding runs on every request by default
2. **Excessive encoder reloading** - creating new tokenizers for each model variant
3. **Long text tokenization** - large prompts/responses dominate CPU time

Both issues are addressed below.

## Solutions Applied

### 1. Reduced Connection Pool Defaults

**File: `model/main.go`**

Changed default connection pool from:
- `SQL_MAX_OPEN_CONNS=1000` → `20`
- `SQL_MAX_IDLE_CONNS=100` → `10` (half of max open)
- Added `SQL_MAX_IDLE_TIME=30` seconds

**Environment Variables** (still fully configurable):
```bash
SQL_MAX_OPEN_CONNS=20        # Max concurrent DB connections
SQL_MAX_IDLE_CONNS=10        # Max idle connections in pool
SQL_MAX_LIFETIME=60          # Max connection lifetime (seconds)
SQL_MAX_IDLE_TIME=30         # Max idle time before closing (seconds)
```

For a 1 vCPU/2 GB server, these are good defaults. Increase only if you have more CPU cores.

### 2. Added Composite Indexes on `logs` Table

**File: `model/log.go`**

Added indexes to speed up common queries:
- `idx_type_created` on `(type, created_at)` - for rpm/tpm rate calculations
- `idx_user_created` on `(user_id, created_at)` - for user log pagination

These indexes will be created automatically on next migration. For existing databases, you can create them manually:

```sql
-- PostgreSQL
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_type_created ON logs (type, created_at);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_created ON logs (user_id, created_at);

-- MySQL
CREATE INDEX idx_type_created ON logs (type, created_at);
CREATE INDEX idx_user_created ON logs (user_id, created_at);
```

**Note:** Use `CONCURRENTLY` in PostgreSQL to avoid blocking writes during index creation.

### 3. Optimized Stats Query (`SumUsedQuota`)

**File: `model/log.go`**

- Separated quota calculation (potentially large time range) from rpm/tpm (last 60 seconds only)
- Added `COALESCE` to prevent NULL issues
- Eliminated redundant WHERE clause duplication

**Before:** Two queries with overlapping conditions  
**After:** Two focused queries with distinct time ranges

### 4. Optimized Token Counting for High RPM

**File: `service/token_counter.go`**

**Problem:** Every request calls `tiktoken` to count tokens, which is CPU-intensive especially for long texts.

**Optimizations applied:**
- **Encoder normalization**: Reuses same tokenizer for model variants (e.g., `gpt-4o-2024-11-20` → `gpt-4o`)
- **Fast estimation for long texts**: Texts > 10k chars use a heuristic (3 chars/token) instead of full tokenization
- **Better caching**: Reduced lock contention on encoder map

**To disable token counting entirely** (if you trust provider-reported usage):
```bash
CountToken=false
```

This eliminates all tokenizer overhead but you'll rely on provider-reported token counts.

**Recommended for high RPM on 1 CPU:**
- Keep `CountToken=true` but ensure texts are reasonably sized
- For very long context windows (>100k tokens), consider disabling or using provider counts
- Monitor CPU: if token counting still dominates, disable it

### 5. PostgreSQL Configuration Recommendations

Add these to your `postgresql.conf` (or via Docker environment variables):

```conf
# For 2 GB RAM server
shared_buffers = 512MB
work_mem = 4MB
maintenance_work_mem = 128MB
effective_cache_size = 1.5GB

# Reduce connection overhead
max_connections = 50

# Track slow queries
log_min_duration_statement = 1000    # Log queries > 1 second
track_io_timing = on
track_functions = all

# Autovacuum tuning (prevent bloat)
autovacuum_max_workers = 2
autovacuum_vacuum_scale_factor = 0.05     # Vacuum when 5% dead tuples
autovacuum_analyze_scale_factor = 0.05
```

For Docker, add to `docker-compose.yml`:

```yaml
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: new_api
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: your_password
    command:
      - "postgres"
      - "-c"
      - "shared_buffers=512MB"
      - "-c"
      - "work_mem=4MB"
      - "-c"
      - "maintenance_work_mem=128MB"
      - "-c"
      - "effective_cache_size=1.5GB"
      - "-c"
      - "max_connections=50"
      - "-c"
      - "log_min_duration_statement=1000"
      - "-c"
      - "track_io_timing=on"
    volumes:
      - postgres_data:/var/lib/postgresql/data
```

### 5. Regular Maintenance

**Delete Old Logs** (reduce table size):

Use the built-in API endpoint or run manually:

```bash
# Via API (admin auth required)
curl -X DELETE "http://your-server/api/log?target_timestamp=1701388800"

# Or SQL
DELETE FROM logs WHERE created_at < extract(epoch from now() - interval '30 days');
VACUUM ANALYZE logs;
```

**Monitor PostgreSQL Activity:**

```sql
-- Check current queries
SELECT pid, state, wait_event_type, query, now() - query_start AS runtime
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY runtime DESC
LIMIT 10;

-- Check top CPU-consuming queries (requires pg_stat_statements)
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

SELECT query, calls, total_exec_time, mean_exec_time, rows
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 10;

-- Check table bloat
SELECT schemaname, tablename, n_dead_tup, n_live_tup, 
       round(100.0 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0), 2) AS dead_ratio
FROM pg_stat_user_tables
WHERE n_dead_tup > 1000
ORDER BY n_dead_tup DESC;
```

## Deployment

1. **Update code** (already done in this repo)
2. **Restart application** - new pool limits take effect immediately
3. **Add indexes** (run SQL above if upgrading existing DB)
4. **Tune PostgreSQL** config as shown above
5. **Monitor** CPU usage over 24-48 hours

## Expected Results

**After database optimizations:**
- **Idle CPU:** 5-15% (down from potential 100%)
- **Active CPU:** 30-60% under load (instead of constant 100%)
- **Query times:** 10-100x faster for log stats and pagination
- **Stable over time:** No gradual degradation over hours/days

**After token counting optimizations:**
- **High RPM CPU:** 20-40% reduction in CPU usage during traffic spikes
- **Long text requests:** Up to 90% faster for texts > 10k characters
- **Encoder overhead:** Near-zero for common model variants (cached)

## Troubleshooting

### CPU still high after hours (database issue)

If CPU still spikes after database optimizations:

1. Check `pg_stat_activity` during spike to identify the query
2. Run `EXPLAIN ANALYZE` on slow queries to verify index usage
3. Check autovacuum status: `SELECT * FROM pg_stat_progress_vacuum;`
4. Consider partitioning the `logs` table by date if > 10M rows
5. Reduce dashboard refresh frequency if stats endpoint is called too often

### CPU spikes during high RPM (request handling)

If CPU usage correlates with request rate:

1. **Check token counting overhead:**
   ```bash
   # Temporarily disable to measure impact
   CountToken=false
   ```
   If CPU drops significantly, token counting is the bottleneck.

2. **Profile the application:**
   - Enable pprof: Set `GIN_MODE=debug` and access `/debug/pprof/profile?seconds=30`
   - Look for `tokenizer.Codec.Count` or `getTokenNum` in CPU profile

3. **Reduce token counting workload:**
   - Trust provider-reported usage: `CountToken=false`
   - Sample counting: Only count tokens for 10% of requests (code change needed)
   - Use async counting: Move token counting to background worker (code change needed)

4. **Optimize request pipeline:**
   - Reduce middleware: Disable unused auth/logging layers
   - Use HTTP/2 or connection pooling on upstream providers
   - Consider horizontal scaling: Add more instances behind load balancer

5. **Hardware upgrade:**
   - 1 vCPU is tight for high RPM; consider 2 vCPU minimum
   - More RAM allows bigger DB cache, reducing DB CPU impact

---

**Last Updated:** December 6, 2025  
**Applies to:** PostgreSQL 12+ / MySQL 5.7+
