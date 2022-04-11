// +build linux, amd64

package persistence

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"database/sql"
	// comment justifying a blank import for the golint
	_ "github.com/lib/pq"
)

// Initialize the package
func Initialize() {
	initMu.Lock()
	defer initMu.Unlock()
	if initialized {
		return
	}
	log.Println("persistence.Initialize")
	// extract variables form command line and environment, env prefered (?)
	flag.Parse()
	if os.Getenv("DBUSER") != "" {
		*dbuser = os.Getenv("DBUSER")
	}
	if os.Getenv("DBPASS") != "" {
		*dbpass = os.Getenv("DBPASS")
	}
	if os.Getenv("DBNAME") != "" {
		*dbname = os.Getenv("DBNAME")
	}
	if os.Getenv("DBHOST") != "" {
		*dbhost = os.Getenv("DBHOST")
	}
	if os.Getenv("DBPORT") != "" {
		if p, err := strconv.Atoi(os.Getenv("DBPORT")); err == nil && p > 0 && p < 65536 {
			*dbport = p
		} else {
			log.Printf("An error parsing DBPORT env. variable: ''%s' => '%v', '%d', Using %d",
				os.Getenv("DBPORT"), err, p, *dbport)
		}
	}
	conn := fmt.Sprintf("user=%s host=%s port=%d dbname=%s password=%s sslmode=disable", // sslmode=require 
		*dbuser, *dbhost, *dbport, *dbname, *dbpass)
	var err error
	db, err = sql.Open("postgres", conn)
	if err != nil {
		db = nil
		log.Printf("Error '%v' opening postgres DB, conn str is '%s'", err, conn)
		return
	}
	initialized = true
	log.Println("Connection Opened ...")
}

// CreateTable - if exists it does nothing
func CreateTable(name string, sql []string) (err error) {
	var success bool
	// create tables if not present
	row := db.QueryRow(`SELECT EXISTS (
		SELECT 1
    FROM   pg_tables
    WHERE  tablename = $1);`, name)
	//log.Printf("create table %s mark 1 => %v", name, err)
	if err = row.Scan(&success); err == nil {
		//log.Printf("create table %s mark 2 => %v", name, success)
		if !success {
			for _, s := range sql {
				if _, err = db.Exec(s); err != nil {
					break
				}
			}
		}
	} else {
		//log.Printf("create table %s mark 3 => %v", name, err)
	}

	return
}

func SizeOfPKey(name string) int {
	rows, err := Query(`SELECT
		  pg_attribute.attname
		FROM pg_index, pg_class, pg_attribute, pg_namespace
		WHERE
		  pg_class.oid = '$1'::regclass AND
		  indrelid = pg_class.oid AND
		  nspname = 'public' AND
		  pg_class.relnamespace = pg_namespace.oid AND
		  pg_attribute.attrelid = pg_class.oid AND
		  pg_attribute.attnum = any(pg_index.indkey)
		 AND indisprimary;`, name)
	if err != nil {
		return -1
	}
	return len(rows)
}

func AddCharColumn(table, column, def string) error {
	return Exec(
		fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s VARCHAR;
			ALTER TABLE %s ALTER COLUMN %s SET DEFAULT '%s';
			UPDATE %s SET %s='%s';
			ALTER TABLE %s ALTER COLUMN %s SET NOT NULL;`,
			table, column, table, column, def, table, column, def, table, column))
}

func DefineNewPKey(table string, column ...string) error {
	err := Exec(fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT $s_pkey;",
		table, table))
	if err != nil {
		return err
	}
	return Exec(fmt.Sprintf("ALTER TABLE %s ADD PRIMARY KEY (%s)",
		table, strings.Join(column, ", ")))
}

// Exec a SQL statement
func Exec(stmt string, params ...interface{}) (err error) {
	if db == nil {
		return errors.New("No DB Connection")
	}
	//log.Println(stmt, params)
	preparedStmt, e := db.Prepare(stmt)
	if e != nil {
		err = e
		return
	}
	//log.Println("statement prepared")
	_, err = preparedStmt.Exec(params...)
	if err != nil {
		log.Printf("Error %v executing %s with %v", err, stmt, params)
	}
	return err
}

// Query SQL
func Query(stmt string, arr ...interface{}) (retval [][]string, err error) {
	preparedStmt, e := db.Prepare(stmt)
	if e != nil {
		err = e
		return
	}
	return StmtQuery(preparedStmt, arr...)
}

// StmtQuery a prepared statement with args, returns array of rows
func StmtQuery(preparedStmt *sql.Stmt, arr ...interface{}) (retval [][]string, err error) {
	rows, e := preparedStmt.Query(arr...)
	if e != nil {
		err = e
		return
	}
	defer rows.Close()
	retval = make([][]string, 0)
	colTypes, e := rows.ColumnTypes()
	if e != nil {
		err = e
		return
	}
	for rows.Next() {
		arr := make([]string, len(colTypes))
		parr := make([]interface{}, len(colTypes))
		for i := range arr {
			parr[i] = &arr[i]
		}
		err = rows.Scan(parr...)
		if err != nil {
			return
		}
		retval = append(retval, arr)
	}
	return
}

// Sqlerr printf SQL error
func Sqlerr(err error) {
	if err != nil {
		log.Printf("SQL Error '%v'", err)
	}
}

// Batch executes same statement with batch of parameters
func Batch(stmt string, arr [][]interface{}) (err error) {
	return BatchInsert(stmt, arr)
}

// BatchInsert https://stackoverflow.com/questions/12486436/golang-how-do-i-batch-sql-statements-with-package-database-sql
func BatchInsert(stmt string, arr [][]interface{}) (err error) {
	if 0 == len(arr) || 0 == len(arr[0]) {
		return nil
	}
	valueStrings := make([]string, 0, len(arr))
	valueArgs := make([]interface{}, 0, len(arr)*len(arr[0]))
	i := 1
	for _, post := range arr {
		questions := "("
		for j := 0; j < len(post); j++ {
			if j > 0 {
				questions += ", "
			}
			questions += fmt.Sprintf("$%d ", i)
			i++
		}
		questions += ")"
		valueStrings = append(valueStrings, questions)
		for _, p := range post {
			valueArgs = append(valueArgs, p)
		}
	}
	stmt += " " + strings.Join(valueStrings, ",") + ";"
	//log.Println(stmt)
	preparedStmt, err := db.Prepare(stmt)
	if err != nil {
		log.Printf("BatchInsert Error %v preparing %s", err, stmt)
	}
	if err != nil {
		return
	}
	//log.Printf("Batch Insert: statement prepared, %d rows\n%s\n%v\n", len(arr), stmt, valueArgs)
	_, err = preparedStmt.Exec(valueArgs...)
	if err != nil {
		log.Printf("BatchInsert Error %v executing %s with %v", err, stmt, arr)
	}
	return
}

// DB returns a pointer to sql.DB
func DB() *sql.DB { return db }
