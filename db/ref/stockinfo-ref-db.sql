BEGIN TRANSACTION;
DROP TABLE IF EXISTS "stockinfo";
CREATE TABLE IF NOT EXISTS "stockinfo" (
	"id"	INTEGER NOT NULL,
	"isin"	TEXT,
	"charturl"	TEXT,
	"name"	TEXT,
	"description"	TEXT,
	"moreinfourl"	TEXT,
	PRIMARY KEY("id" AUTOINCREMENT)
);
DROP TABLE IF EXISTS "operation";
CREATE TABLE IF NOT EXISTS "operation" (
	"id"	INTEGER NOT NULL,
	"unit"	INTEGER NOT NULL,
	"priceunit"	REAL,
	"pricetotal"	REAL,
	"isin"	TEXT,
	"timestamp"	INTEGER,
	PRIMARY KEY("id" AUTOINCREMENT)
);
DROP TABLE IF EXISTS "secrets";
CREATE TABLE IF NOT EXISTS "secrets" (
	"id"	INTEGER NOT NULL,
	"clientid"	TEXT,
	"clientsecret"	TEXT,
	"authtoken"	TEXT,
	"refreshtoken"	TEXT,
	"email"	TEXT,
	"accesstoken"	TEXT,
	PRIMARY KEY("id" AUTOINCREMENT)
);
COMMIT;
