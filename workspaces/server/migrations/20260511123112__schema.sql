-- Modify "schedules" table
ALTER TABLE "public"."schedules" ALTER COLUMN "pattern" TYPE jsonb USING pattern::jsonb;
