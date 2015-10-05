package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	go_n1ql "github.com/couchbaselabs/go_n1ql"
	"log"
)

var serverURL = flag.String("server", "http://localhost:9000",
	"couchbase server URL")

func main() {

	flag.Parse()
	n1ql, err := sql.Open("n1ql", *serverURL)
	if err != nil {
		log.Fatal(err)
	}

	err = n1ql.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Set query parameters
	//os.Setenv("n1ql_timeout", "10s")
	ac := []byte(`[{"user": "admin:Administrator", "pass": "password"}]`)
	//os.Setenv("n1ql_creds", string(ac))

	go_n1ql.SetQueryParams("creds", string(ac))
	go_n1ql.SetQueryParams("timeout", "10s")

	/*
		result, err := n1ql.Exec("Create primary index on `beer-sample`")
		if err != nil {
			log.Fatal(err)
		}
	*/

	name := "brewery"
	rows, err := n1ql.Query("select * from `beer-sample` where type = ?", name)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	rowsReturned := 0
	for rows.Next() {
		var contacts string
		if err := rows.Scan(&contacts); err != nil {
			log.Fatal(err)
		}
		rowsReturned++
	}

	log.Printf("Rows returned %d : \n", rowsReturned)
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	stmt, err := n1ql.Prepare("Upsert INTO default values (?,?)")
	if err != nil {
		log.Fatal(err)
	}
	// Map Values need to be marshaled
	value, _ := json.Marshal(map[string]interface{}{"name": "irish", "type": "contact"})

	var rowsAffected int64
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("irish%d", i)
		value := make(map[string]interface{})
		value["name"] = key
		value["type"] = "contact"
		valueBytes, _ := json.Marshal(value)
		result, err := stmt.Exec(key, valueBytes)

		if err != nil {
			fmt.Errorf(" Failed here %v", err)
			continue
		}

		ra, err := result.RowsAffected()
		if err != nil {
			log.Fatal(err)
		}
		rowsAffected += ra
	}

	log.Printf("Total Rows Affected %d", rowsAffected)

	stmt.Close()
	_, err = stmt.Exec("test", "this shouldn't work")
	if err == nil {
		log.Fatal("Statement not closed")
	}

	_, err = n1ql.Exec("delete from default  use keys ? ", "irish")
	if err != nil {
		log.Fatal(err)
	}

	keys := make([]string, 0)
	for i := 0; i < 100000; i++ {
		keys = append(keys, fmt.Sprintf("irish%d", i))
	}

	value, _ = json.Marshal(keys)
	_, err = n1ql.Exec("delete from default use keys ?", value)
	if err != nil {
		log.Fatal(err)
	}

}
