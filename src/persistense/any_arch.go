package persistence

import (
	"database/sql"
	"flag"
	"sync"
)

var (
	initMu sync.Mutex

	dbuser = flag.String("dbu", "postgres", "DB username")                 // or DB_USER
	dbpass = flag.String("dbpsw", "hello", "DB Password")                       // or DB_PSW
	dbname = flag.String("dbn", "postgres", "DB Name")                                         // or DB_NAME
	dbhost = flag.String("dbh", "db", "DB Host") // or DB_HOST
	dbport = flag.Int("dbport", 5432, "DB Port")                                          // or DB_PORT

	db          *sql.DB
	initialized = false
)

// Tier returns DB tier
func Tier() string { return *dbname }

// DBUser returns db user (for debug purposes in main)
func DBUser() string { return *dbuser }

// DBPass returns db psw (for debug purposes in main)
func DBPass() string { return *dbpass }

// SetDBPass sets db user (for debug purposes in main)
func SetDBPass(psw string) { *dbpass = psw }

// DBName returns db name (for debug purposes in main)
func DBName() string { return *dbname }

// DBHost returns db host (for debug purposes in main)
func DBHost() string { return *dbhost }

// DBPort returns db port (for debug purposes in main)
func DBPort() int { return *dbport }
