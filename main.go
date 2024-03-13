package main

import (
	"fmt"
	"log"
)

func main() {
	config := DefaultConfig()
	db, err := New(config)
	if err != nil {
		log.Fatal(err)
	}
	t, _ := db.Begin(true)
	err = t.Set("hello", "world")
	if err != nil {
		fmt.Printf("%s", err)
	}
	t.Commit()
	value, err := t.get("hello")
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Printf("%s", value)
	db.Close()
	fmt.Printf("done")
	defer db.Close()
}
