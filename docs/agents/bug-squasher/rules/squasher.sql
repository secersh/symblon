SELECT e.actor_login
FROM events e,
     unnest(json_extract(e.payload, '$.issue.labels')::JSON[]) AS t(label)
WHERE e.event_type  = 'issues'
  AND e.action      = 'closed'
  AND json_extract_string(label, '$.name') = 'bug'
  AND e.received_at BETWEEN :window_start AND :window_end
GROUP BY e.actor_login
HAVING COUNT(*) >= 5
