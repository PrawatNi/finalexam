package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Customer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

type CustomerInfo struct {
	ID     int
	Name   string
	Email  string
	Status string
}

//DB Pointer
var db *sql.DB

//var stmt *sql.Stmt
//var rows *sql.Rows
//var err error

func connectDB() {
	url := os.Getenv("DATABASE_URL")
	fmt.Println("url :", url)
	var err error
	db, err = sql.Open("postgres", url)
	if err != nil {
		log.Println("Connect to database error", err)
		return
	}
}

//insert One Customer
func postOneCustomerHandler(c *gin.Context) {
	var customer Customer
	err := c.ShouldBindJSON(&customer)
	if err != nil {
		log.Println("can't Bind Body to JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't Bind Body to JSON : " + err.Error()})
		return
	}

	row := db.QueryRow("INSERT INTO customers (name, email, status) values ($1, $2, $3 ) RETURNING id", customer.Name, customer.Email, customer.Status)
	err = row.Scan(&customer.ID)
	if err != nil {
		log.Println("can't scan : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "can't scan : " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, customer)
}

//select All Customer
func getAllCustomerHandler(c *gin.Context) {
	var customers []Customer
	var customerOne Customer
	stmt, err := db.Prepare("SELECT id, name, email, status From customers")
	if err != nil {
		log.Println("can't prepare query statement : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "can't prepare query statement : " + err.Error()})
		return
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Println("can't query statement : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "can't query statement : " + err.Error()})
		return
	}
	for rows.Next() {
		err := rows.Scan(&customerOne.ID, &customerOne.Name, &customerOne.Email, &customerOne.Status)
		if err != nil {
			log.Println("can't scan rows : ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "can't scan rows : " + err.Error()})
			return
		}
		customers = append(customers, customerOne)
	}
	c.JSON(http.StatusOK, customers)
}

//select One Customer
func getOneCustomerHandler(c *gin.Context) {
	var customerOne Customer
	stmt, err := db.Prepare("SELECT id, name, email, status From customers where id=$1")
	if err != nil {
		log.Println("can't prepare query one customer statement : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "can't prepare query customer statement : " + err.Error()})
		return
	}
	rowId := c.Param("id")
	row := stmt.QueryRow(rowId)
	err = row.Scan(&customerOne.ID, &customerOne.Name, &customerOne.Email, &customerOne.Status)
	if err != nil {
		log.Println("can't Scan row into variables : ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't prepare query one statement : " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, customerOne)
}

//update One Customer
func updateOneCustomerHandler(c *gin.Context) {
	var customerOneInfo CustomerInfo
	var customerOne Customer

	err := c.ShouldBindJSON(&customerOneInfo)
	if err != nil {
		log.Println("can't Bind Body to JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't Bind Body to JSON : " + err.Error()})
		return
	}
	stmt, err := db.Prepare("UPDATE customers SET name=$2, email=$3, status=$4 WHERE id=$1")
	if err != nil {
		log.Println("can't prepare statement update : ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't prepare update statement: " + err.Error()})
		return
	}
	if _, err := stmt.Exec(customerOneInfo.ID, customerOneInfo.Name, customerOneInfo.Email, customerOneInfo.Status); err != nil {
		log.Println("error execute update", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't execute update statement: " + err.Error()})
		return
	}

	customerOne.ID = strconv.Itoa(customerOneInfo.ID)
	customerOne.Name = customerOneInfo.Name
	customerOne.Email = customerOneInfo.Email
	customerOne.Status = customerOneInfo.Status

	c.JSON(http.StatusOK, customerOne)
}

//delete One Customer
func deleteOneCustomerHandler(c *gin.Context) {
	rowId := c.Param("id")
	stmt, err := db.Prepare("DELETE FROM customers WHERE id=$1")

	numId, err := strconv.Atoi(rowId)
	if err != nil {
		log.Println("can't convert to int -", rowId, "- : ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't convert to int -" + rowId + "- : " + err.Error()})
		return
	}
	_, err = stmt.Query(numId)
	if err != nil {
		log.Println("can't Querty : ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "can't execContext : " + err.Error()})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
}

//create Table
func createCustomerTable() {
	createTable := `
	CREATE TABLE IF NOT EXISTS customers (
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		status TEXT
	);
	`
	_, err := db.Exec(createTable)
	if err != nil {
		log.Println("can't create table : ", err)
		return
	}
}

func authMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token != "token2019" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized."})
		c.Abort()
		return
	}
	c.Next()
}

func main() {
	/*
		POST /customers
		GET /customers/:id
		GET /customers
		PUT /customers/:id
		DELETE /customers/:id
	*/
	connectDB()
	createCustomerTable()
	r := gin.Default()
	r.Use(authMiddleware)
	r.POST("/customers", postOneCustomerHandler)
	r.GET("/customers", getAllCustomerHandler)
	r.GET("/customers/:id", getOneCustomerHandler)
	r.PUT("/customers/:id", updateOneCustomerHandler)
	r.DELETE("/customers/:id", deleteOneCustomerHandler)
	r.Run(":2019") // listen and serve on 0.0.0.0:2019
	defer db.Close()
}
