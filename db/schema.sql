CREATE TYPE transaction_status AS ENUM ('CREATED', 'IN_PROGRESS', 'CANCELLED', 'FAILED', 'SUCCESS');

CREATE TABLE IF NOT EXISTS "user" (
  "id" SERIAL PRIMARY KEY,
  "login" varchar(256)
);

CREATE TABLE IF NOT EXISTS "transaction" (
  "id" SERIAL PRIMARY KEY,
  "timestamp" timestamp,
  "from_balance" int,
  "to_balance" int,
  "value" numeric(19, 4),
  "currency" varchar,
  "status" transaction_status DEFAULT 'CREATED'
);

CREATE TABLE IF NOT EXISTS "currency" (
  "code" varchar PRIMARY KEY,
  "name" varchar
);

CREATE TABLE IF NOT EXISTS "balance" (
  "id" SERIAL PRIMARY KEY,
  "user" int,
  "currency" varchar,
  "balance" numeric(19, 4)
);

ALTER TABLE "user" ADD CONSTRAINT unique_login UNIQUE ("login");

ALTER TABLE "transaction" ADD FOREIGN KEY ("from_balance") REFERENCES "balance" ("id");

ALTER TABLE "transaction" ADD FOREIGN KEY ("to_balance") REFERENCES "balance" ("id");

ALTER TABLE "transaction" ADD FOREIGN KEY ("currency") REFERENCES "currency" ("code");

ALTER TABLE "balance" ADD FOREIGN KEY ("user") REFERENCES "user" ("id");

ALTER TABLE "balance" ADD FOREIGN KEY ("currency") REFERENCES "currency" ("code");
