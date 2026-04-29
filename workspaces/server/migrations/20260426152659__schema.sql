-- Create "tokens" table
CREATE TABLE "public"."tokens" (
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "id" text NOT NULL,
  "token" text NOT NULL,
  "user_id" text NOT NULL,
  "type" text NOT NULL,
  "expires" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_token" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
