package main

import (
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"time"
)

func main() {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{"127.0.0.1:9000"},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
		//TLS: &tls.Config{
		//	InsecureSkipVerify: true,
		//},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 150 * time.Second,
		//Compression: &clickhouse.Compression{
		//	clickhouse.CompressionLZ4,
		//},
		Debug: true,
	})
	conn.SetMaxIdleConns(5)
	conn.SetMaxOpenConns(10)
	conn.SetConnMaxLifetime(time.Hour)
	//defer conn.Close()

	err := conn.Ping()
	if err != nil {
		fmt.Printf("Ping error: %s\n", err)
	}

	//var res *int32
	row := conn.QueryRow("SELECT * FROM t1 limit 1")
	fmt.Printf("%v", row.Err())
	//err = row.Scan(&res)
	//fmt.Printf("%v", err)

	//var res2 string
	//row = conn.QueryRow("SELECT * FROM t2")
	//row.Scan(&res2)

	//fmt.Printf("%v\n", *res)
}
