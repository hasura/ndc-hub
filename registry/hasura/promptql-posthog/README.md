# PostHog Connector for Hasura DDN

A Hasura NDC (Native Data Connector) that enables querying PostHog analytics via HogQL queries, discovering dashboards, and retrieving insight definitions.

## Features

- Execute HogQL queries (ClickHouse SQL dialect) against your PostHog data
- List all dashboards with metadata (timestamps, pinned status, insight summaries)
- Retrieve detailed insight definitions including query/filter configurations
- Ideal for AI agents to discover analytics topics and problems in an account

## Environment Variables

| Variable | Required | Description | Default |
|----------|----------|-------------|---------|
| `POSTHOG_API_KEY` | Yes | PostHog API key for authentication. Get from PostHog Settings > Project > Personal API Keys | - |
| `POSTHOG_PROJECT_ID` | Yes | PostHog project ID. Found in PostHog URL: `app.posthog.com/project/{PROJECT_ID}` | - |
| `POSTHOG_HOST` | No | PostHog API host URL | `us.posthog.com` |

## HogQL Query Examples

### Count events by type
```sql
SELECT event, count() as count
FROM events
WHERE timestamp > now() - INTERVAL 7 DAY
GROUP BY event
ORDER BY count DESC
LIMIT 10
```

### Get unique users per day
```sql
SELECT
  toDate(timestamp) as date,
  uniq(distinct_id) as unique_users
FROM events
WHERE timestamp > now() - INTERVAL 30 DAY
GROUP BY date
ORDER BY date
```

### Page view funnel
```sql
SELECT
  event,
  count() as count
FROM events
WHERE event IN ('$pageview', 'sign_up', 'purchase')
  AND timestamp > now() - INTERVAL 7 DAY
GROUP BY event
```

### User properties
```sql
SELECT
  properties.$browser as browser,
  count() as count
FROM events
WHERE timestamp > now() - INTERVAL 7 DAY
GROUP BY browser
ORDER BY count DESC
```

## PostHog Host URLs

| Region | Host |
|--------|------|
| US Cloud | `us.posthog.com` |
| EU Cloud | `eu.posthog.com` |
| Self-hosted | Your PostHog instance URL |
