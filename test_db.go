package main

import (
	"fmt"
	"log"

	"github.com/elithrar/simple-scrypt"
	"github.com/tidwall/buntdb"
)

var db *buntdb.DB

func main() {
	// Open the data.db file. It will be created if it doesn't exist.
	db, err := buntdb.Open("data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// e.g. r.PostFormValue("password")
	passwordFromForm := "enrico"

	// Generates a derived key of the form "N$r$p$salt$dk" where N, r and p are defined as per
	// Colin Percival's scrypt paper: http://www.tarsnap.com/scrypt/scrypt.pdf
	// scrypt.Defaults (N=16384, r=8, p=1) makes it easy to provide these parameters, and
	// (should you wish) provide your own values via the scrypt.Params type.
	hash, err := scrypt.GenerateFromPassword([]byte(passwordFromForm), scrypt.DefaultParams)
	if err != nil {
		log.Fatal(err)
	}

	// Print the derived key with its parameters prepended.
	fmt.Printf("%s\n", hash)

	//Set a value
	err = db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("user:vytek", string(hash), nil)
		return err
	})

	err = db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get("user:vytek")
		if err != nil {
			return err
		}
		fmt.Printf("value is %s\n", val)
		return nil
	})

	err = db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("", func(key, value string) bool {
			fmt.Printf("key: %s, value: %s\n", key, value)
			return true // continue iteration
		})
		return err
	})
}
