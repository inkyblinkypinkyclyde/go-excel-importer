package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	_ "github.com/lib/pq"
)

func connectdb(pgUser string, pgPassword string) *sql.DB {
	connStr := "postgresql://localhost/importedExcelSheets?user=" + pgUser + "&password=" + pgPassword + "&sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err)
	}
	return db
}
func main() {
	var documentName string
	var pgUser string
	var pgPassword string
	fmt.Println("Enter your postgres username: ")
	fmt.Scanln(&pgUser)
	fmt.Println("Enter your postgres password: ")
	fmt.Scanln(&pgPassword)
	fmt.Println("Enter the name of the excel document: ")
	fmt.Scanln(&documentName)
	started := time.Now()

	f, err := excelize.OpenFile(documentName + ".xlsx")
	if err != nil {
		fmt.Println(err)
		fmt.Printf("Error opening file")
		return
	}

	firstSheet := f.WorkBook.Sheets.Sheet[0].Name

	rows := f.GetRows(firstSheet)

	fmt.Printf("Total rows: %d\n", len(rows))

	db := connectdb(pgUser, pgPassword)

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

	_, err = db.Exec("CREATE TABLE " + documentName + " (id SERIAL PRIMARY KEY)")
	if err != nil {
		fmt.Println("\nError creating table", err)
	}

	for _, column := range columnHeaders {
		_, err := db.Exec("ALTER TABLE " + documentName + " ADD COLUMN " + column + " VARCHAR(255)")
		if err != nil {
			fmt.Println("\nError adding column: ", err)
			fmt.Printf("Column: %s", column)
		}
	}

	id := 1
	for _, row := range rows[1:] {
		columnIndex := 0
		_, err := db.Exec("INSERT INTO "+documentName+" (id) VALUES ($1)", id)
		if err != nil {
			fmt.Println("\nError inserting id: ", err)
		}
		for _, cell := range row {

			idforSQL := strconv.Itoa(id)
			sqlStatement := "UPDATE " + documentName + " SET " + columnHeaders[columnIndex] + " = '" + cell + "' WHERE id = " + idforSQL
			_, err := db.Exec(sqlStatement)
			if err != nil {
				fmt.Println("\nError inserting data: ", err)
				fmt.Printf("Cell: %s, Column: %s, ID: %d", cell, columnHeaders[columnIndex], id)
			}
			columnIndex++
		}
		id++
	}

	db.Close()
	finished := time.Now()
	fmt.Printf("Script took %s to run \n", finished.Sub(started))

}
