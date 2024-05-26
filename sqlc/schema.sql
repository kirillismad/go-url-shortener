CREATE TABLE IF NOT EXISTS "links" (
	"id" bigint GENERATED ALWAYS AS IDENTITY NOT NULL UNIQUE,
	"short_id" text NOT NULL UNIQUE,
	"href" text NOT NULL UNIQUE,
	"created_at" timestamp with time zone NOT NULL DEFAULT NOW(),
	"usage_count" bigint NOT NULL DEFAULT 0,
	"usage_at" timestamp with time zone NOT NULL DEFAULT NOW(),
	PRIMARY KEY ("id")
);