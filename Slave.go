package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"
	"time" //only

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

func init() {
	godotenv.Load()

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/school", user, pass)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
}

func ExecuteLocal(query string) {
	query = strings.TrimSpace(query)
	if strings.HasPrefix(strings.ToUpper(query), "SELECT") {
		rows, err := db.Query(query)
		if err != nil {
			fmt.Println("Local SELECT error:", err)
			return
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		values := make([]interface{}, len(cols))
		for i := range values {
			var v interface{}
			values[i] = &v
		}

		for rows.Next() {
			rows.Scan(values...)
			for i, col := range values {
				fmt.Printf("%s: %v\t", cols[i], *(col.(*interface{})))
			}
			fmt.Println()
		}
	} else {
		_, err := db.Exec(query)
		if err != nil {
			fmt.Println("Local execution error:", err)
		}
	}
}

func ConnectToMaster() {
	for {
		conn, err := net.Dial("tcp", "100:8080")
		if err != nil {
			fmt.Println("Master unavailable. Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("Connected to master")
		defer conn.Close()

		go ReceiveFromMaster(conn)

		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter SQL query: ")
			query, _ := reader.ReadString('\n')
			query = strings.TrimSpace(query)

			if !isValidSQL(query) {
				fmt.Println("Query not allowed from slave terminal. Only master can issue critical operations.")
				continue
			}

			_, err := conn.Write([]byte(query))
			if err != nil {
				fmt.Println("Connection lost. Trying to reconnect...")
				break
			}
		}
	}
}

func isValidSQL(query string) bool {
	query = strings.Join(strings.Fields(strings.ToUpper(query)), " ")

	if strings.HasPrefix(query, "CREATE DATABASE") ||
		strings.HasPrefix(query, "DROP DATABASE") ||
		strings.HasPrefix(query, "CREATE TABLE") ||
		strings.HasPrefix(query, "ALTER TABLE") ||
		strings.HasPrefix(query, "DROP TABLE") {
		return false
	}

	return strings.HasPrefix(query, "SELECT") ||
		strings.HasPrefix(query, "INSERT") ||
		strings.HasPrefix(query, "UPDATE") ||
		strings.HasPrefix(query, "DELETE")
}

// ReceiveFromMaster listens to master's commands and executes them (even critical ones)
func ReceiveFromMaster(conn net.Conn) {
	trustMaster := true // allow execution of critical queries if from master

	buffer := make([]byte, 4096)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Disconnected from master")
			return
		}
		query := string(buffer[:n])
		query = strings.TrimSpace(query)

		fmt.Println("\nReceived from master:", query)

		if trustMaster || isValidSQL(query) {
			ExecuteLocal(query)
		} else {
			fmt.Println("Blocked query from master (disallowed by policy):", query)
		}
	}
}

func main() {
	ConnectToMaster()
}
