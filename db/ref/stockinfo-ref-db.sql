BEGIN TRANSACTION;
DROP TABLE IF EXISTS "secrets";
CREATE TABLE IF NOT EXISTS "secrets" (
	"id"	INTEGER NOT NULL,
	"clientid"	TEXT,
	"clientsecret"	TEXT,
	"authtoken"	TEXT,
	"refreshtoken"	TEXT,
	PRIMARY KEY("id" AUTOINCREMENT)
);
DROP TABLE IF EXISTS "stockinfo";
CREATE TABLE IF NOT EXISTS "stockinfo" (
	"id"	INTEGER NOT NULL,
	"isin"	TEXT,
	"charturl"	TEXT,
	"name"	TEXT,
	"description"	TEXT,
	PRIMARY KEY("id" AUTOINCREMENT)
);
COMMIT;
