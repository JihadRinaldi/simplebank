ALTER TABLE "accounts" DROP CONSTRAINT "owner_currency_key";

ALTER TABLE "accounts" DROP CONSTRAINT "accounts_owner_fkey";

DROP TABLE IF EXISTS "users";