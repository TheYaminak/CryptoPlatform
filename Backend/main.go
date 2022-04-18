package main

import (
	"fmt"
	"log"
	"net/http"
	"proyecto/router"

	"github.com/rs/cors"

	_ "github.com/lib/pq"
)

func main() {
	r := router.Router()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)
	fmt.Println("Starting server on the port 3020...")
	log.Fatal(http.ListenAndServe(":3020", handler))
}
