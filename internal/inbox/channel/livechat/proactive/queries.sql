-- name: history
SELECT id, campaign_id, inbox_id, browser_key, session_key, COALESCE(contact_id, 0) AS contact_id,
       snapshot, created_at, displayed, opened, dismissed, replied, COALESCE(conversation_uuid::text, '') AS conversation_uuid
FROM widget_campaign_deliveries
WHERE inbox_id = $1 AND (browser_key = $2 OR (contact_id = $3 AND $3 > 0))
ORDER BY created_at DESC;

-- name: reserve
INSERT INTO widget_campaign_deliveries(campaign_id, inbox_id, browser_key, session_key, contact_id, snapshot, url, mobile)
VALUES ($1, $2, $3, $4, NULLIF($5, 0), $6, $7, $8)
RETURNING id;

-- name: get-delivery
SELECT id, campaign_id, inbox_id, browser_key, session_key, COALESCE(contact_id, 0) AS contact_id,
       snapshot, created_at, displayed, opened, dismissed, replied, COALESCE(conversation_uuid::text, '') AS conversation_uuid
FROM widget_campaign_deliveries
WHERE id = $1 AND inbox_id = $2
  AND ((contact_id = $4 AND $4 > 0) OR (browser_key = $3 AND
       (contact_id IS NULL OR ($4 = 0 AND EXISTS (SELECT 1 FROM users WHERE users.id = contact_id AND users.type = 'visitor')))));

-- name: record-event
UPDATE widget_campaign_deliveries
SET displayed = displayed OR $2 = 'displayed',
    opened = opened OR $2 = 'opened',
    dismissed = dismissed OR $2 = 'dismissed'
WHERE id = $1;

-- name: bind-contact
UPDATE widget_campaign_deliveries SET contact_id = $3
WHERE inbox_id = $1 AND browser_key = $2 AND contact_id IS NULL;

-- name: stats
SELECT campaign_id,
       COUNT(*) FILTER (WHERE displayed)::int AS displayed,
       COUNT(*) FILTER (WHERE opened)::int AS opened,
       COUNT(*) FILTER (WHERE dismissed)::int AS dismissed,
       COUNT(*) FILTER (WHERE replied)::int AS replied
FROM widget_campaign_deliveries
WHERE inbox_id = $1 AND created_at >= $2 AND created_at < $3
GROUP BY campaign_id;

-- name: deliveries
SELECT d.id, d.campaign_id, d.created_at, d.url, d.mobile,
       COALESCE(d.snapshot->>'message', '') AS message,
       COALESCE(d.snapshot->>'sender', '') AS sender,
       d.displayed, d.opened, d.dismissed, d.replied,
       COALESCE(d.conversation_uuid::text, '') AS conversation_uuid,
       COALESCE(d.contact_id, 0) AS contact_id,
       COALESCE(u.type::text, '') AS contact_type,
       TRIM(CONCAT(COALESCE(u.first_name, ''), ' ', COALESCE(u.last_name, ''))) AS contact_name,
       COALESCE(u.email, '') AS contact_email,
       COUNT(*) OVER() AS total
FROM widget_campaign_deliveries d
LEFT JOIN users u ON u.id = d.contact_id
WHERE d.inbox_id = $1 AND d.created_at >= $2 AND d.created_at < $3
  AND ($4 = '' OR d.campaign_id::text = $4)
  AND ($5 = ''
       OR ($5 = 'displayed' AND d.displayed)
       OR ($5 = 'undisplayed' AND NOT d.displayed)
       OR ($5 = 'opened' AND d.opened)
       OR ($5 = 'dismissed' AND d.dismissed)
       OR ($5 = 'replied' AND d.replied))
  AND ($6 = '' OR POSITION($6 IN d.url) > 0)
ORDER BY d.created_at DESC, d.id
LIMIT $7 OFFSET $8;
