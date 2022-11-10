package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	_ "github.com/lib/pq"
)

// func createdb() {
// 	db, err := sql.Open(

func connectdb() *sql.DB {
	connStr := "postgresql://localhost/importedExcelSheets?user=richardgannon&password=postgres&sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err)
	}
	return db
}
func main() {
	started := time.Now()
	// fmt.Printf("Started script at %s \n", started)
	f, err := excelize.OpenFile("EmployeeSampleData.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}

	rows := f.GetRows("Data")

	// for _, row := range rows {
	// 	fmt.Println(row)
	// }
	fmt.Printf("Total rows: %d", len(rows))
	fmt.Printf("rows is a %s", reflect.TypeOf(rows))
	// connect to the database
	db := connectdb()
	// get the column headers from the excel file and remove spaces
	ColumnHeadersWithSpaces := rows[0]
	var columnHeaders []string
	for _, column := range ColumnHeadersWithSpaces {
		column = strings.Replace(column, " ", "_", -1)
		column = strings.Replace(column, "/", "_", -1)
		column = strings.Replace(column, "(", "", -1)
		column = strings.Replace(column, ")", "", -1)
		column = strings.Replace(column, "-", "_", -1)
		column = strings.Replace(column, ".", "", -1)
		column = strings.Replace(column, ":", "", -1)
		column = strings.Replace(column, "%", "", -1)
		columnHeaders = append(columnHeaders, column)
	}
	// create the table
	_, err = db.Exec("CREATE TABLE employees (id SERIAL PRIMARY KEY)") // TODO: change this to a variable
	if err != nil {
		fmt.Println("\nError creating table", err)
	}
	// add the columns to the table
	for _, column := range columnHeaders {
		_, err := db.Exec("ALTER TABLE employees ADD COLUMN " + column + " VARCHAR(255)") // TODO: change this to a variable
		if err != nil {
			fmt.Println("\nError adding column: ", err)
			fmt.Printf("Column: %s", column)
		}
	}

	// insert the data into the table
	id := 1
	for _, row := range rows[1:] {
		columnIndex := 0
		_, err := db.Exec("INSERT INTO employees (id) VALUES ($1)", id)
		if err != nil {
			fmt.Println("\nError inserting id: ", err)
		}
		for _, cell := range row {
			cell = strings.Replace(cell, " ", "_", -1)
			cell = strings.Replace(cell, "/", "-", -1)
			cell = strings.Replace(cell, "(", "", -1)
			cell = strings.Replace(cell, ")", "", -1)
			cell = strings.Replace(cell, ".", "", -1)
			cell = strings.Replace(cell, ":", "", -1)
			cell = strings.Replace(cell, "%", "", -1)
			idforSQL := strconv.Itoa(id)
			sqlStatement := "UPDATE employees SET " + columnHeaders[columnIndex] + " = '" + cell + "' WHERE id = " + idforSQL
			fmt.Println(sqlStatement)
			_, err := db.Exec(sqlStatement) // TODO: change this to a variable
			if err != nil {
				fmt.Println("\nError inserting data: ", err)
				fmt.Printf("Cell: %s, Column: %s, ID: %d", cell, columnHeaders[columnIndex], id)
			}
			columnIndex++
		}
		id++
	}
	// close the connection
	// check that the data is in there

	db.Close()
	finished := time.Now()
	// fmt.Printf("\nFinished script at %s \n", finished)
	fmt.Printf("Script took %s to run \n", finished.Sub(started))

}
