SELECT DISTINCT e.actor_login
FROM events e,
     unnest(json_extract(e.payload, '$.issue.labels')::JSON[]) AS t(label)
WHERE e.event_type = 'issues'
  AND e.action     = 'closed'
  AND json_extract_string(label, '$.name') = 'bug'
  AND NOT EXISTS (
      SELECT 1
      FROM issued_symbols s
      WHERE s.actor_login = e.actor_login
        AND s.symbol_id   = 'first-blood'
  )
