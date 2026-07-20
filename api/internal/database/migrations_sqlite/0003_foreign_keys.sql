-- SQLite doesn't support easy ALTER TABLE ADD CONSTRAINT FOREIGN KEY.
-- Since the application layer already prevents taxonomy deletion via ErrTaxonomyInUse
-- and the Go code manually cleans up post_tags on post update, we can skip SQLite FKs here
-- to avoid full table rebuilds.
SELECT 1;
