-- Modify "occurances" table
ALTER TABLE "public"."occurances" ALTER COLUMN "schedule_id" SET NOT NULL, ALTER COLUMN "start" SET NOT NULL, ALTER COLUMN "end" SET NOT NULL;
-- Create index "idx_occurances_schedule_id" to table: "occurances"
CREATE INDEX "idx_occurances_schedule_id" ON "public"."occurances" ("schedule_id");
-- Modify "rooms" table
ALTER TABLE "public"."rooms" ALTER COLUMN "code" SET NOT NULL, ALTER COLUMN "name" SET NOT NULL, ALTER COLUMN "host_id" SET NOT NULL, ALTER COLUMN "url" SET NOT NULL;
-- Modify "schedules" table
ALTER TABLE "public"."schedules" ALTER COLUMN "room_id" SET NOT NULL, ALTER COLUMN "start" SET NOT NULL, ALTER COLUMN "end" SET NOT NULL, ALTER COLUMN "pattern" SET NOT NULL;
-- Create index "idx_schedules_room_id" to table: "schedules"
CREATE INDEX "idx_schedules_room_id" ON "public"."schedules" ("room_id");
-- Modify "shorteners" table
ALTER TABLE "public"."shorteners" ALTER COLUMN "code" TYPE character varying(16), ALTER COLUMN "code" SET NOT NULL, ALTER COLUMN "url" SET NOT NULL;
