package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"net"//TCP protocol
	"os"
	"strings"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"//load .env (environment variables)
)

var db *sql.DB
var slaveConns []net.Conn
/*
db: Database connection object.

slaveConns: Slice to hold all TCP connections from slave clients.
*/
func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")

	if user == "" || pass == "" {
		fmt.Println("Environment variables not loaded. DB_USER or DB_PASS is empty.")
		os.Exit(1)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/school", user, pass)
	fmt.Println("Using DSN:", dsn)//DSN data source name

	var errDB error
	db, errDB = sql.Open("mysql", dsn)
	if errDB != nil {
		panic(errDB)
	}
}

func StartTCPServer() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Master listening on port 8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		slaveConns = append(slaveConns, conn)
		go HandleSlave(conn)
	}
}

func HandleSlave(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Read identity line (e.g., [GUI] or [SLAVE])
	identityLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Failed to read identity.")
		return
	}
	identityLine = strings.TrimSpace(identityLine)

	for {
		query, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected.")
			return
		}
		query = strings.TrimSpace(query)

		if identityLine != "[GUI]" && isCriticalQuery(query) {
			errMsg := "Error: Critical queries like DROP or CREATE are not allowed from slaves.\n"
			conn.Write([]byte(errMsg))
			fmt.Printf("Blocked critical query from slave %v: %s\n", conn.RemoteAddr(), query)
			return
		}

		result := ExecuteQuery(query)
		conn.Write([]byte(result))
	}
}



func isCriticalQuery(query string) bool {
	query = strings.ToUpper(query)
	return strings.HasPrefix(query, "DROP DATABASE") ||
		strings.HasPrefix(query, "CREATE DATABASE") ||
		strings.HasPrefix(query, "CREATE TABLE")
}

func ExecuteQuery(query string) string {
	query = strings.TrimSpace(query)
	if strings.HasPrefix(strings.ToUpper(query), "SELECT") {
		rows, err := db.Query(query)
		if err != nil {
			return "SELECT Error: " + err.Error()
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		var result strings.Builder
		result.WriteString(strings.Join(cols, "\t") + "\n")

		for rows.Next() {
			rows.Scan(valuePtrs...)
			for _, val := range values {
				if val == nil {
					result.WriteString("NULL\t")
				} else {
					switch v := val.(type) {
					case []byte:
						result.WriteString(fmt.Sprintf("%s\t", string(v)))
					default:
						result.WriteString(fmt.Sprintf("%v\t", v))
					}
				}
			}
			result.WriteString("\n")
		}
		return result.String()
	} else {
		_, err := db.Exec(query)
		if err != nil {
			return "Error: " + err.Error()
		}
		return "Success"
	}
}

func main() {
	go StartTCPServer()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter SQL query: ")
		query, _ := reader.ReadString('\n')
		query = strings.TrimSpace(query)
		result := ExecuteQuery(query)
		fmt.Println("Master:", result)

		if isCriticalQuery(query) {
			fmt.Println("Executed critical query on master only.")
		} else {
			for _, slave := range slaveConns {
				slave.Write([]byte(query))
			}
		}
		
	}
}
