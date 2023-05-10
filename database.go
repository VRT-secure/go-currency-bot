package main

import (
    "database/sql"
    "fmt"
	"strings"
    _ "github.com/mattn/go-sqlite3"
)


func createDB(table_name, columns_with_types string) *sql.DB {
	database, err := sql.Open("sqlite3", "./currencyes_price.db")
	if err != nil {
		panic(err)
	}
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", table_name, columns_with_types)
    statement, err := database.Prepare(query)
	if err != nil {
		panic(err)
	}
    statement.Exec()
	return database
}


func insertDB(db *sql.DB, table_name string, columns, values []string) error {
	// Создаем строку плейсхолдеров
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	
	// Создаем SQL запрос
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table_name,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	// Теперь вы можете подготовить и выполнить запрос
	stmt, err := db.Prepare(query)

	// Преобразуем значения в []interface{} для Exec
	valuesInterface := make([]interface{}, len(values))
	for i, v := range values {
		valuesInterface[i] = v
	}

	_, err = stmt.Exec(valuesInterface...)
	
	return err
}


func selectDB(db *sql.DB)  {
	// rows, err := db.Query("SELECT * FROM table")

	// defer rows.Close()

	// Получаем имена столбцов
	// columns, err := rows.Columns()
	// if err != nil {
		
	// }

	
}

// TODO - закончить функцию или начать использовать https://github.com/go-gorm/gorm
