-- Postgres migrations
ALTER TABLE post_tags DROP CONSTRAINT IF EXISTS post_tags_post_id_fkey;
ALTER TABLE post_tags ADD CONSTRAINT post_tags_post_id_fkey FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE;

ALTER TABLE post_tags DROP CONSTRAINT IF EXISTS post_tags_tag_id_fkey;
ALTER TABLE post_tags ADD CONSTRAINT post_tags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE RESTRICT;

ALTER TABLE posts DROP CONSTRAINT IF EXISTS posts_category_id_fkey;
ALTER TABLE posts ADD CONSTRAINT posts_category_id_fkey FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT;

-- Add same for comments and replies if not already
ALTER TABLE comments DROP CONSTRAINT IF EXISTS comments_post_slug_fkey;
ALTER TABLE comments ADD CONSTRAINT comments_post_slug_fkey FOREIGN KEY (post_slug) REFERENCES posts(slug) ON DELETE CASCADE;
