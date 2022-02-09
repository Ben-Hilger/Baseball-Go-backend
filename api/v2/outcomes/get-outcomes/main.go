package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
)

type Outcome struct {
	PitchOutcomeID      int    `json:"pitch_outcome_id"`
	OutcomeShortName    string `json:"short_name"`
	OutcomeLongName     string `json:"long_name"`
	StrikesToAdd        int    `json:"strikes_to_add"`
	BallsToAdd          int    `json:"balls_to_add"`
	OutsToAdd           int    `json:"outs_to_add"`
	StopAfterTwoStrikes bool   `json:"stop_after_two_strikes"`
}

type Result struct {
	Data []Outcome `json:"data"`
}

func setupDB() *sql.DB {
	db_user := os.Getenv("DB_USER")
	db_name := os.Getenv("DB_NAME")
	db_pass := os.Getenv("DB_PASSWORD")
	db_port := os.Getenv("DB_PORT")
	db_host := os.Getenv("DB_HOST")
	db, err := sql.Open("mysql",
		db_user+":"+db_pass+"@tcp("+db_host+":"+db_port+")/"+db_name)
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Getting available hands...")
	// Configure the database connection
	db := setupDB()
	// Query the player data
	rows, err := db.Query("CALL GetOutcomes();")
	// Check the query for issues
	if err != nil {
		log.Panicln(err)
	}
	// Store the returned data
	var data []Outcome
	// Iterate through all of the rows
	for rows.Next() {
		// Store the next sequence
		result := Outcome{}
		s := reflect.ValueOf(&result).Elem()
		numCols := s.NumField()
		columns := make([]interface{}, numCols)
		for i := 0; i < numCols; i++ {
			field := s.Field(i)
			columns[i] = field.Addr().Interface()
		}
		err := rows.Scan(columns...)
		if err != nil {
			log.Println(err)
		}
		data = append(data, result)
	}
	res, err := json.Marshal(Result{Data: data})
	if err != nil {
		log.Println(err)
	}
	log.Println("Got all available hands...")
	// Return the result
	return events.APIGatewayProxyResponse{Body: string(res), StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
