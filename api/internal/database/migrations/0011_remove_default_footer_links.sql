UPDATE operation_documents
SET data = jsonb_set(data, '{footerItems}', '[]'::jsonb, true)
WHERE key = 'navigation'
  AND jsonb_typeof(data->'footerItems') = 'array'
  AND jsonb_array_length(data->'footerItems') = 3
  AND data->'footerItems' @> '[
    {"id":"nav_footer_1","url":"/"},
    {"id":"nav_footer_2","url":"/archive"},
    {"id":"nav_footer_3","url":"/topics"}
  ]'::jsonb;
