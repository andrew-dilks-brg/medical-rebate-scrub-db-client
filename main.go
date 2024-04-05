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
)

// prod database
// const (
// 	host     = "localhost"
// 	port     = 10008
// 	user     = "postgres"
// 	password = "1234"
// 	dbname   = "citus"
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

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
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
	rows, err := db.Query(fmt.Sprintf("DELETE FROM medical_rebate_scrub.%s WHERE manu='%s'", table, manu))
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
	itemsAdded := 0

	if table == "ndc_lu" || table == "hcpcs_lu" {
		file = convertCSVToMap(fileLocation, true, false) // need swap true to handle map primary key
		for key, value := range file {
			itemsAdded += addNdcItem(db, table, manu, key, value, debug)
		}
	} else if table == "pos_lu" {
		file = convertCSVToMap(fileLocation, false, false) // no swap needed
		for key := range file {
			itemsAdded += addPosItem(db, table, manu, key, debug)
		}
	} else if table == "mod_lu" {
		file = convertCSVToMap(fileLocation, false, false) // no swap needed
		for key := range file {
			itemsAdded += addModItem(db, table, manu, key, debug)
		}
	} else if table == "csr_list" {
		results := parseCSRcsv(fileLocation, false) // different parser entirely
		for _, item := range results {
			itemsAdded += addCsrItem(db, table, manu, item, debug)
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
		_, err := db.Query(fmt.Sprintf("INSERT INTO medical_rebate_scrub.%s VALUES ('%s', '%s', '%s');", table, manu, value, key))
		if err != nil {
			fmt.Println(err.Error() + " for item: " + key + ", " + value)
		} else {
			itemsAdded += 1
		}
	}
	return itemsAdded
}

func addPosItem(db *sql.DB, table string, manu string, key string, debug bool) int {
	itemsAdded := 0
	if !strings.Contains(key, "POS") {
		if debug { fmt.Println(key) }
		_, err := db.Query(fmt.Sprintf("INSERT INTO medical_rebate_scrub.%s VALUES ('%s', '%s');", table, manu, key))
		if err != nil {
			fmt.Println(err.Error() + " for item: " + key)
		} else {
			itemsAdded += 1
		}
	}
	return itemsAdded
}

func addModItem(db *sql.DB, table string, manu string, key string, debug bool) int {
	itemsAdded := 0
	if !strings.Contains(key, "MOD_340B") {
		if debug { fmt.Println(key) }
		_, err := db.Query(fmt.Sprintf("INSERT INTO medical_rebate_scrub.%s VALUES ('%s', '%s');", table, manu, key))
		if err != nil {
			fmt.Println(err.Error() + " for item: " + key)
		} else {
			itemsAdded += 1
		}
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
		on, err := db.Query( fmt.Sprintf("INSERT INTO medical_rebate_scrub.%s VALUES ('%s', '%s', '%s', '%s', '%s', '%s');", table, manu, csrNPI, csrProductName, csrVal, csrStartDOS, csrEndDOS) )
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
		rows, _ = db.Query(fmt.Sprintf("SELECT * FROM medical_rebate_scrub.%s", table))
	} else {
		rows, _ = db.Query(fmt.Sprintf("SELECT * FROM medical_rebate_scrub.%s WHERE manu='%s'", table, manu))
	}
	parseTableOutput(table, rows)
	fmt.Println()
}

func printTableHeaders(table string) {
	if table == "ndc_lu" {
		fmt.Println("manu  | product |  ndc")
	} else if table == "pos_lu" {
		fmt.Println("manu  |  pos")
	} else if table == "mod_lu" {
		fmt.Println("manu  |  mod_340b")
	} else if table == "hcpcs_lu" {
		fmt.Println("manu | product | hcpcs_cd")
	} else if table == "csr_list" {
		fmt.Println("manu | npi | product | csr | start_date | term_date")
	} else {
		fmt.Println("You havent built support for this table yet")
	}
}

func parseTableOutput(table string, rows *sql.Rows) {
	if table == "ndc_lu" {
		var manufacturer string
		var product string
		var ndc string
		for rows.Next() {
			rows.Scan(&manufacturer, &product, &ndc)
			fmt.Println("" + manufacturer + ", " + product + ", " + ndc)
		}
	} else if table == "pos_lu" {
		var manufacturer string
		var pos string
		for rows.Next() {
			rows.Scan(&manufacturer, &pos)
			fmt.Println("" + manufacturer + ", " + pos)
		}
	} else if table == "mod_lu" {
		var manufacturer string
		var mod_340b string
		for rows.Next() {
			rows.Scan(&manufacturer, &mod_340b)
			fmt.Println("" + manufacturer + ", " + mod_340b)
		}
	} else if table == "hcpcs_lu" {
		var manufacturer string
		var product string
		var hcpcs_cd string
		for rows.Next() {
			rows.Scan(&manufacturer, &product, &hcpcs_cd)
			fmt.Println("" + manufacturer + ", " + product + ", " + hcpcs_cd)
		}
	} else if table == "csr_list" {
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
	} else {
		fmt.Println("You havent built support for this table yet")
	}
}