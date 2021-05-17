package db

const (
	// SchemaVersion is the version of the SQLite schema understood by this backend.
	// Whenever changes to the schema are made, this version number should be increased
	// to prevent data corruption between incompatible database schemas.
	SchemaVersion = 2

	// Schema is the SQLite database layout.
	Schema = `
	CREATE TABLE "zones" (
		"id"	INTEGER NOT NULL,
		"name"	TEXT NOT NULL,
		PRIMARY KEY("id")
	);

	CREATE TABLE "sites" (
		"tp_address"	TEXT NOT NULL,
		"name"	TEXT NOT NULL,
		PRIMARY KEY("tp_address")
	);

	CREATE TABLE subnets(
		net_ip BLOB NOT NULL,
		net_mask BLOB NOT NULL,
		zone INTEGER NOT NULL,
		tp_address TEXT NOT NULL,
		PRIMARY KEY (net_ip, net_mask),
		FOREIGN KEY (zone) REFERENCES Zones(id) ON DELETE CASCADE,
		FOREIGN KEY (tp_address) REFERENCES Sites(tp_address) ON DELETE CASCADE
	  );
	  
	CREATE TABLE "transitions" (
		"policyID"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"src"	INTEGER,
		"dest"	INTEGER,
		"srcPort"	INTEGER,
		"destPort"	INTEGER,
		"proto"	TEXT,
		"action"	TEXT NOT NULL,
		FOREIGN KEY("src") REFERENCES "Zones"("id") ON DELETE CASCADE,
		FOREIGN KEY("dest") REFERENCES "Zones"("id") ON DELETE CASCADE
	);`

	ZonesTable       = "Zones"
	SitesTable       = "Sites"
	SubnetsTable     = "Subnets"
	TransitionsTable = "Transitions"
)
