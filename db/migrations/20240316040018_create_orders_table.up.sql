CREATE TABLE IF NOT EXISTS orders (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  product_id UUID NOT NULL,
  bank_account_id UUID NOT NULL,
  payment_proof_image_url VARCHAR(128) NOT NULL,
  quantity INTEGER NOT NULL,
  created_at TIMESTAMP(0) DEFAULT NOW(),
  updated_at TIMESTAMP(0) DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_product_id ON orders(product_id);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
