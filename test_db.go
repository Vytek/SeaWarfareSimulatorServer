package main

import (
	"log"
	"fmt"

	"github.com/tidwall/buntdb"
)

func main() {
	// Open the data.db file. It will be created if it doesn't exist.
	db, err := buntdb.Open("data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//Set a value
	err = db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("mykey", "myvalue", nil)
		return err
	})

	err = db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get("mykey")
		if err != nil{
			return err
		}
		fmt.Printf("value is %s\n", val)
		return nil
	})
}