-- Modify "users" table
ALTER TABLE "public"."users" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE text, ADD COLUMN "created_at" timestamptz NULL, ADD COLUMN "updated_at" timestamptz NULL, ADD COLUMN "deleted_at" timestamptz NULL, ADD COLUMN "verified" boolean NULL DEFAULT false, ADD COLUMN "password" text NULL, ADD COLUMN "profile_image_url" text NULL;
-- Drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "public"."users_id_seq";
-- Create index "idx_users_deleted_at" to table: "users"
CREATE INDEX "idx_users_deleted_at" ON "public"."users" ("deleted_at");
