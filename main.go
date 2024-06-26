package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

// Staging database
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "1234"
	dbname   = "Adilks"
	sslEnabled = false
	schemaName = "rbtbin"
)

// prod database
// const (
// 	host     = "c.ovi-ledger.postgres.database.azure.com"
// 	port     = 6432
// 	user     = "citus"
// 	password = "GET FROM ADMIN"
// 	dbname   = "citus"
// 	sslEnabled = true
// 	schemaName = "rbtbin"
// )

func missingRequiredArgs() {
	fmt.Println("Missing required args! Exiting")
	fmt.Println("USAGE: go run . -manu=MANU -step=STEP -file=FILE.csv")
	fmt.Println("USAGE: STEP can be 'add'|'delete'|'get' ")

	os.Exit(1)
}

// README - manu must be a valid manufacturer - case insensitive
// step must be the process you'd like to run for example add get, or delete
// file path required for add
// TODO - currently schema is hardcoded for table headers and printing table output - this shouldnt change but already has once - if happens again look into using a DB DAO
func main() {
	m := flag.String("manu", "", "")
	s := flag.String("step", "", "")
	t := flag.String("table", "", "")
	f := flag.String("file", "", "")
	flag.Parse()

	debug := false
	manufacturer := strings.ToLower(*m)
	step := strings.ToLower(*s)
	fileLocation := *f
	table := *t
	fmt.Println("Running with args: manu=" + manufacturer + ", step=" + step + ", file=" + fileLocation + ", table=" + table)

	if step == "get" || step == "delete" {
		// dont need file
		if table == "" {
			missingRequiredArgs()
		}
	} else if step == "add" {
		if manufacturer == "" || fileLocation == "" || table == "" {
			missingRequiredArgs()
		}
	} else {
		missingRequiredArgs()
	}

	var psqlInfo string
	if !sslEnabled {
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",host, port, user, password, dbname)
	} else {
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",host, port, user, password, dbname)
	}

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Println()
		fmt.Println("Was not able to ping the DB!")
		fmt.Println()
		panic(err)
	}

	if step=="add" {
		queryTable(db, table, "")
		addItems(db, table, manufacturer, fileLocation, debug)
	} else if step == "delete" {
		queryTable(db, table, "")
		deleteItems(db, table, manufacturer)
	} else if step == "get" {
		queryTable(db, table, manufacturer)
	}
}

func deleteItems(db *sql.DB, table string, manu string) {
	fmt.Println("- Deleting data from table - ")
	rows, err := db.Query(fmt.Sprintf("DELETE FROM %s.%s WHERE manu='%s'", schemaName, table, manu))
	if err != nil {
		fmt.Println(err)
	} else {
		for rows.Next() {
			var output string
			fmt.Println(rows.Scan(&output))
		}
	}
	fmt.Println("")
	queryTable(db, table, manu)
}

func addItems(db *sql.DB, table string, manu string, fileLocation string, debug bool) {
	fmt.Println("- Adding data into table - ")
	var file map[string]string
	var results []map[string]string
	itemsAdded := 0

	if table == "mrs_ndc_lu" || table == "mrs_hcpcs_lu" {
		file = convertCSVToMap(fileLocation, true, false) // need swap true to handle map primary key
		for key, value := range file {
			itemsAdded += addNdcItem(db, table, manu, key, value, debug)
		}
	} else if table == "mrs_pos_lu" {
		results = parseMultiColcsv(fileLocation, false)
		for _, key := range results {
			itemsAdded += addPosItem(db, table, manu, key["POS"], key["POS_TYPE"], debug)
		}
	} else if table == "mrs_mod_lu" {
		results = parseMultiColcsv(fileLocation, false)
		for _, key := range results {
			itemsAdded += addModItem(db, table, manu, key["MOD_340B"], key["MOD_TYPE"], debug)
		}
	} else if table == "mrs_csr_list" {
		results = parseMultiColcsv(fileLocation, false)
		for _, item := range results {
			itemsAdded += addCsrItem(db, table, manu, item, debug)
		}
	} else if table == "mrs_binary_cbks" {
		results = parseMultiColcsv(fileLocation, false)
		for _, item := range results {
			itemsAdded += addBinaryCbksItem(db, table, manu, item, debug)
		}
	} else {
		fmt.Println("Have yet to add support for this table!")
	}
	fmt.Println()
	
	if itemsAdded > 0 {
		queryTable(db, table, manu)
		fmt.Println("Successfully added " + strconv.Itoa(itemsAdded) + " items!")
	} else {
		fmt.Println("Nothing to do!")
	}
	fmt.Println()
}

func addNdcItem(db *sql.DB, table string, manu string, key string, value string, debug bool) int {
	itemsAdded := 0
	if !strings.Contains(value, "PRODUCT") {
		if debug { fmt.Println(key + ", " + value) }
		_, err := db.Query(fmt.Sprintf("INSERT INTO %s.%s VALUES ('%s', '%s', '%s');", schemaName, table, manu, value, key))
		if err != nil {
			fmt.Println(err.Error() + " for item: " + key + ", " + value)
		} else {
			itemsAdded += 1
		}
	}
	return itemsAdded
}

func addPosItem(db *sql.DB, table string, manu string, pos string, pos_type string, debug bool) int {
	itemsAdded := 0
	if debug { fmt.Println(pos) }
	_, err := db.Query(fmt.Sprintf("INSERT INTO %s.%s VALUES ('%s', '%s', '%s');", schemaName, table, manu, pos, pos_type))
	if err != nil {
		fmt.Println(err.Error() + " for item: " + pos)
	} else {
		itemsAdded += 1
	}

	return itemsAdded
}

func addModItem(db *sql.DB, table string, manu string, mod_340b string, mod_type string, debug bool) int {
	itemsAdded := 0
	if debug { fmt.Println(mod_340b) }
	_, err := db.Query(fmt.Sprintf("INSERT INTO %s.%s VALUES ('%s', '%s', '%s');", schemaName, table, manu, mod_340b, mod_type))
	if err != nil {
		fmt.Println(err.Error() + " for item: " + mod_340b)
	} else {
		itemsAdded += 1
	}
	return itemsAdded
}

func addBinaryCbksItem(db *sql.DB, table string, manu string, item map[string]string, debug bool) int {
	itemsAdded := 0
	cbkProductName := item["PRODUCT"]
	cbkNPI         := item["NPI"]
	cbkDesc        := item["DESCRIPTION"]
	cbkPriority    := item["PRIORITY"]

	if debug { fmt.Println(item) }
	on, err := db.Query( fmt.Sprintf(
		"INSERT INTO %s.%s VALUES ('%s', '%s', '%s', '%s', '%s');", schemaName, table, manu, cbkProductName, cbkNPI, cbkDesc, cbkPriority,
	) )

	if err != nil {
		fmt.Println(err.Error() + " for item: " + cbkProductName + " " + cbkNPI)
	} else {
		on.Close()
		itemsAdded += 1
	}

	return itemsAdded
}

func addCsrItem(db *sql.DB, table string, manu string, item map[string]string, debug bool) int {
	itemsAdded := 0
	csrNPI         := item["NPI"]
	csrVal         := item["CSR"]
	csrProductName := item["PRODUCT"]
	csrStartDOS    := item["START_DATE"]
	csrEndDOS      := item["TERM_DATE"]

	if !strings.Contains(csrNPI, "MOD_340B") {
		if debug { fmt.Println(item) }
		on, err := db.Query( fmt.Sprintf("INSERT INTO %s.%s VALUES ('%s', '%s', '%s', '%s', '%s', '%s');", schemaName, table, manu, csrNPI, csrProductName, csrVal, csrStartDOS, csrEndDOS) )
		// pq: sorry, too many clients already
		if err != nil {
			fmt.Println(err.Error() + " for item: " + csrNPI + " " + csrProductName)
		} else {
			on.Close()
			itemsAdded += 1
		}
	}
	return itemsAdded
}

func queryTable(db *sql.DB, table string, manu string) {
	var rows *sql.Rows

	fmt.Println("- Querying table -")
	printTableHeaders(table)
	if manu == "" {
		rows, _ = db.Query(fmt.Sprintf("SELECT * FROM %s.%s", schemaName, table))
	} else {
		rows, _ = db.Query(fmt.Sprintf("SELECT * FROM %s.%s WHERE manufacturer='%s'", schemaName, table, manu))
	}
	parseTableOutput(table, rows)
	fmt.Println()
}

func printTableHeaders(table string) {
	if table == "mrs_ndc_lu" {
		fmt.Println("manufacturer  | product |  ndc")
	} else if table == "mrs_pos_lu" {
		fmt.Println("manufacturer  |  pos  |  type")
	} else if table == "mrs_mod_lu" {
		fmt.Println("manufacturer  |  mod_340b  |  mod_type")
	} else if table == "mrs_hcpcs_lu" {
		fmt.Println("manufacturer | product | hcpcs_cd")
	} else if table == "mrs_csr_list" {
		fmt.Println("manufacturer | npi | product | csr | start_date | term_date")
	} else if table == "mrs_binary_cbks" {
		fmt.Println("manufacturer | product | npi | description | priority")
	} else {
		fmt.Println("You havent built support for this table yet")
	}
}

func parseTableOutput(table string, rows *sql.Rows) {
	if table == "mrs_ndc_lu" {
		var manufacturer string
		var product string
		var ndc string
		for rows.Next() {
			rows.Scan(&manufacturer, &product, &ndc)
			fmt.Println("" + manufacturer + ", " + product + ", " + ndc)
		}
	} else if table == "mrs_pos_lu" {
		var manufacturer string
		var pos string
		var pos_type string
		for rows.Next() {
			rows.Scan(&manufacturer, &pos, &pos_type)
			fmt.Println("" + manufacturer + ", " + pos + ", " + pos_type)
		}
	} else if table == "mrs_mod_lu" {
		var manufacturer string
		var mod_340b string
		var mod_type string
		for rows.Next() {
			rows.Scan(&manufacturer, &mod_340b, &mod_type)
			fmt.Println("" + manufacturer + ", " + mod_340b + ", " + mod_type)
		}
	} else if table == "mrs_hcpcs_lu" {
		var manufacturer string
		var product string
		var hcpcs_cd string
		for rows.Next() {
			rows.Scan(&manufacturer, &product, &hcpcs_cd)
			fmt.Println("" + manufacturer + ", " + product + ", " + hcpcs_cd)
		}
	} else if table == "mrs_csr_list" {
		var manufacturer string
		var npi string
		var product string
		var csr string
		var start_date string
		var term_date string
		for rows.Next() {
			rows.Scan(&manufacturer, &npi, &product, &csr, &start_date, &term_date)
			fmt.Println("" + manufacturer + ", " + npi + ", " + product + ", " + csr + ", " + start_date + ", " + term_date)
		}
	} else if table == "mrs_binary_cbks" {
		var manufacturer string
		var product string
		var npi string
		var description string
		var priority string
		for rows.Next() {
			rows.Scan(&manufacturer, &product, &npi, &description, &priority)
			fmt.Println("" + manufacturer + ", " + product + ", " + npi + ", " + description + ", " + priority)
		}
	} else {
		fmt.Println("You havent built support for this table yet")
	}
}
