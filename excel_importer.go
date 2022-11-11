package main

import (
	"database/sql"
	"fmt"

	// "net/http"
	// "reflect"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type ExcelData struct {
	pgUser     string
	pgPassword string
	pgTable    string
	db         *sql.DB
}

var documentName string = "EmployeeSampleDataFiveRows"
var pgUser string = "richardgannon"
var pgPassword string = "postgres"

var dbValues = &ExcelData{pgUser, pgPassword, documentName, connectdb(pgUser, pgPassword)}

func getData(c *gin.Context) {
	column := c.Param("column")
	datum := c.Param("datum")
	// var columnNames []string
	// sql_query := "SELECT * FROM " + dbValues.pgTable + " WHERE " + column + " = '" + datum + "';"
	sql_query := "SELECT * FROM " + dbValues.pgTable
	fmt.Println(sql_query)
	rows, err := dbValues.db.Query(sql_query)
	if err != nil {
		fmt.Println("Error getting data: ", err)
	}
	// columnNames, err := rows.Columns()
	columns, err := rows.Columns()
	if err != nil {
		fmt.Println("Error getting columns: ", err)
	}
	// data := make(map[string]string)
	for rows.Next() {
		fmt.Println(columns)
		// data[columns[0]] = rows.Scan(&data[columns[0]])

	}
	fmt.Println(dbValues.db)
	fmt.Printf("Column: %s, Datum: %s", column, datum)
	// fmt.Printf("db is a : %s", reflect.TypeOf(db))
	// c.IndentedJSON(http.StatusOK, ) // data is going to be data pulled from the database
}

func connectdb(pgUser string, pgPassword string) *sql.DB {

	connStr := "postgresql://localhost/importedExcelSheets?user=" + pgUser + "&password=" + pgPassword + "&sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err)
	}
	return db
}
func main() {

	// dbValues := &ExcelData{pgUser, pgPassword, "importedExcelSheets", connectdb(pgUser, pgPassword)}
	// fmt.Println("Enter your postgres username: ")
	// fmt.Scanln(&pgUser)
	// fmt.Println("Enter your postgres password: ")
	// fmt.Scanln(&pgPassword)
	// fmt.Println("Enter the name of the excel document: ")
	// fmt.Scanln(&documentName)
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
			cell = strings.Replace(cell, " ", "_", -1)
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
	router := gin.Default()
	router.GET(":column/:datum", getData)
	router.Run(":8080")

}
