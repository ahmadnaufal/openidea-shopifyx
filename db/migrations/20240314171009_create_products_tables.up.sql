CREATE TABLE IF NOT EXISTS products (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  name VARCHAR(62) NOT NULL,
  price INTEGER NOT NULL,
  img_url VARCHAR(255) NOT NULL,
  stock INTEGER NOT NULL DEFAULT 0,
  condition VARCHAR(16) NOT NULL DEFAULT 'used',
  is_purchasable BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP(0) DEFAULT NOW(),
  updated_at TIMESTAMP(0) DEFAULT NOW()
);

CREATE SEQUENCE product_tags_id_seq;

CREATE TABLE IF NOT EXISTS product_tags (
  id INTEGER NOT NULL DEFAULT NEXTVAL('product_tags_id_seq'),
  product_id UUID NOT NULL,
  tag VARCHAR(32) NOT NULL,
  created_at TIMESTAMP(0) DEFAULT NOW()
);