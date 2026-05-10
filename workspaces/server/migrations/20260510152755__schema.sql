-- Create "shorteners" table
CREATE TABLE "public"."shorteners" (
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "id" bigserial NOT NULL,
  "code" character varying(8) NULL,
  "url" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_shorteners_code" to table: "shorteners"
CREATE UNIQUE INDEX "idx_shorteners_code" ON "public"."shorteners" ("code", "code");
-- Create index "idx_shorteners_deleted_at" to table: "shorteners"
CREATE INDEX "idx_shorteners_deleted_at" ON "public"."shorteners" ("deleted_at");
-- Create "rooms" table
CREATE TABLE "public"."rooms" (
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "id" bigserial NOT NULL,
  "code" character varying(8) NULL,
  "name" text NULL,
  "description" text NULL,
  "host_id" text NULL,
  "url" text NULL,
  "pin" text NULL,
  "auto_join" boolean NULL DEFAULT true,
  "allow_guests" boolean NULL DEFAULT false,
  "allow_recording" boolean NULL DEFAULT false,
  "allow_chat" boolean NULL DEFAULT true,
  "allow_screen_share" boolean NULL DEFAULT true,
  "capacity" bigint NULL DEFAULT 10,
  "location" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_rooms_host" FOREIGN KEY ("host_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_rooms_code" to table: "rooms"
CREATE UNIQUE INDEX "idx_rooms_code" ON "public"."rooms" ("code");
-- Create index "idx_rooms_deleted_at" to table: "rooms"
CREATE INDEX "idx_rooms_deleted_at" ON "public"."rooms" ("deleted_at");
-- Create "schedules" table
CREATE TABLE "public"."schedules" (
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "id" bigserial NOT NULL,
  "room_id" bigint NULL,
  "start" timestamptz NULL,
  "end" timestamptz NULL,
  "pattern" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_rooms_schedules" FOREIGN KEY ("room_id") REFERENCES "public"."rooms" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_schedules_deleted_at" to table: "schedules"
CREATE INDEX "idx_schedules_deleted_at" ON "public"."schedules" ("deleted_at");
-- Create "occurances" table
CREATE TABLE "public"."occurances" (
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "id" bigserial NOT NULL,
  "schedule_id" bigint NULL,
  "start" timestamptz NULL,
  "end" timestamptz NULL,
  "is_cancelled" boolean NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_occurances_schedule" FOREIGN KEY ("schedule_id") REFERENCES "public"."schedules" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_occurances_deleted_at" to table: "occurances"
CREATE INDEX "idx_occurances_deleted_at" ON "public"."occurances" ("deleted_at");
