package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"


	"github.com/gin-gonic/gin"
	stripe "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/webhook"
)

var db *sql.DB


type User struct {
  ID int64 
  Name string
  Email string
  Credit int64 
  ClientReferenceID string
}


func handleWebhookRoute (c *gin.Context) {
  const MaxBodyBytes = int64(65536)

  reqBody := http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

  body, err := io.ReadAll(reqBody)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
    c.AbortWithStatus(http.StatusServiceUnavailable)
    return
  }

  // read request body that was sent from Stripe
  // ?client_reference_id=123

  endpointSecret:= os.Getenv("endpointSecret")

  if endpointSecret == "" {
    fmt.Errorf("set env endpointSecret")
    c.AbortWithStatus(http.StatusServiceUnavailable)
    return
  }

  event, err := webhook.ConstructEvent(body, c.Request.Header.Get("Stripe-Signature"), endpointSecret)

  if err != nil {
    fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
    c.AbortWithStatus(http.StatusBadRequest) // Return a 400 error on a bad signature
    return 

  }

 // Handle the checkout.session.completed event
  if event.Type == "checkout.session.completed" {
    var session stripe.CheckoutSession
    err := json.Unmarshal(event.Data.Raw, &session)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
      c.AbortWithStatus(http.StatusBadRequest)
      return
    }

    if session.PaymentStatus == "paid" {
      if session.ClientReferenceID == "" {
        // can try to look up customer email provided in DB
        fmt.Errorf("empty client_reference_id")
        c.AbortWithStatusJSON(http.StatusBadRequest, "empty client_reference_id")
        return
      }

      fmt.Println("client_reference_id: ", session.ClientReferenceID)
      fmt.Println("the whole event look like this: ", session)
      FulfillOrder(session.ClientReferenceID, session.AmountTotal)
    }

    } 

  c.IndentedJSON(http.StatusOK, event)

  }

func FulfillOrder(customer_id string, amount_total int64) {
 fmt.Println("increase their credit here: ", customer_id)
  var credit int64

  if amount_total >= 5400 {
    credit = amount_total / 450
  } else {
    credit = amount_total / 900
  }
  
  // Open a connection to the database
  var err error
	db, err = sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal("failed to open db connection", err)
	}

  UpdateCredit(customer_id, credit)
  
}

func UpdateCredit(customer_id string, credit_amount int64) {
  credit := GetCurrentCredit(customer_id)
  credit = credit + credit_amount * 100
  
	query := `UPDATE users SET credit = ? WHERE client_reference_id = ?`
  _, err := db.Exec(query, credit, customer_id)
	if err != nil {
		log.Fatal("(UpdateProduct) db.Exec", err)
	}
}

func GetCurrentCredit(customer_id string) int64 {
  var user User

  query := `SELECT * FROM users WHERE client_reference_id = ?`
  err := db.QueryRow(query, customer_id).Scan(&user.ID, &user.Name, &user.Email, &user.Credit, &user.ClientReferenceID)
	if err != nil {
		log.Fatal("(GetCurrentCredit) db.Exec", err)
	}

  return user.Credit
}
