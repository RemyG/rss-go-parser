CREATE TABLE rss_category (
  id serial primary key,
  name varchar(255) NOT NULL,
  cat_order integer NOT NULL DEFAULT 0
);

CREATE TABLE rss_feed (
  id serial primary key,
  link varchar(255) NOT NULL,
  base_link varchar(255) DEFAULT NULL,
  title varchar(255) NOT NULL,
  description text,
  updated timestamp DEFAULT NULL,
  to_update boolean NOT NULL DEFAULT true,
  mark_new_to_read boolean NOT NULL DEFAULT false,
  category_id int NOT NULL,
  valid boolean NOT NULL DEFAULT true,
  viewframe boolean NOT NULL DEFAULT false,
  cat_order integer NOT NULL DEFAULT 0
);

CREATE TABLE rss_entry (
  id bigserial PRIMARY KEY,
  published timestamp DEFAULT NULL,
  updated timestamp DEFAULT NULL,
  link varchar(255) NOT NULL,
  title varchar(255) NOT NULL,
  description text,
  author varchar(255) DEFAULT NULL,
  read boolean NOT NULL default false,
  content text,
  feed_id int NOT NULL,
  favourite boolean NOT NULL DEFAULT false,
  to_read boolean NOT NULL DEFAULT false
);