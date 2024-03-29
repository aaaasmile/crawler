BEGIN TRANSACTION;
DROP TABLE IF EXISTS "price";
CREATE TABLE IF NOT EXISTS "price" (
	"id"	INTEGER NOT NULL,
	"price"	REAL,
	"timestamp"	INTEGER,
	"idstock"	INTEGER NOT NULL,
	FOREIGN KEY("idstock") REFERENCES "stockinfo"("id"),
	PRIMARY KEY("id")
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
	"relaymail"	TEXT,
	"realaysecret"	TEXT,
	"relayhost"	TEXT,
	"relayuser"	TEXT,
	PRIMARY KEY("id" AUTOINCREMENT)
);
DROP TABLE IF EXISTS "stockinfo";
CREATE TABLE IF NOT EXISTS "stockinfo" (
	"id"	INTEGER NOT NULL,
	"isin"	TEXT,
	"charturl"	TEXT,
	"name"	TEXT,
	"description"	TEXT,
	"moreinfourl"	TEXT,
	"cost"	REAL,
	"quantity"	REAL,
	"simple"	TEXT,
	"disabled"	INTEGER NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT)
);
COMMIT;
