package main

import (
	"fmt"
	"net/http"
)

func main() {

	fmt.Println("Server Started on Port :8080")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Server Error ", err)
	}
}
