package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)


func main() {
  err := godotenv.Load()
	if err != nil {
    fmt.Printf("Error: %f",errors.New("no .env file found"))
	}

  port := os.Getenv("PORT")

  if port == "" {
    log.Println("Cannot find PORT env, default to run on port 8080 instead")
    port = "8080"
  }
  router := gin.Default()

  router.POST("/webhook", handleWebhookRoute)  

  
  // router.Use(TokenAuthMiddleware)


  if err := router.Run(":"+port); err != nil {
    log.Fatalf("Couldnot run the server %v", err)
  }
}

func TokenAuthMiddleware(c *gin.Context) {
  token := c.Query("token")
  
  if token == "" {
    c.AbortWithStatusJSON(http.StatusForbidden, "Unauthorize")
  }

  c.Next()
}

