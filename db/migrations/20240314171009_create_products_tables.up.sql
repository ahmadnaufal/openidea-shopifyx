CREATE TABLE IF NOT EXISTS products (
  id VARCHAR(64) PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL,
  name VARCHAR(62) NOT NULL,
  price INTEGER NOT NULL,
  image_url VARCHAR(255) NOT NULL,
  stock INTEGER NOT NULL DEFAULT 0,
  condition VARCHAR(16) NOT NULL DEFAULT 'second',
  is_purchasable BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP(0) DEFAULT NOW(),
  updated_at TIMESTAMP(0) DEFAULT NOW(),
  deleted_at TIMESTAMP(0)
);

CREATE SEQUENCE product_tags_id_seq;

CREATE TABLE IF NOT EXISTS product_tags (
  id INTEGER NOT NULL DEFAULT NEXTVAL('product_tags_id_seq'),
  product_id VARCHAR(64) NOT NULL,
  tag VARCHAR(32) NOT NULL,
  created_at TIMESTAMP(0) DEFAULT NOW()
);
