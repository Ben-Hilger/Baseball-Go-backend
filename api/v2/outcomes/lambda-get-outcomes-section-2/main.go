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

type ExtraInfo struct {
	PitchExtraInfoID   int    `json:"pitch_extra_info_id"`
	ShortName          string `json:"short_name"`
	LongName           string `json:"long_name"`
	MarkHitterAsOut    bool   `json:"mark_hitter_as_out"`
	SendHitterToBaseID int    `json:"send_hitter_to_base_id"`
	CanRunnerAdvance   bool   `json:"can_runner_advance"`
}

var (
	outcomeIDTag = os.Getenv("OUTCOME_ID_TAG")
)

type Result struct {
	Data []ExtraInfo `json:"data"`
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
	log.Println("Getting section 1 options for given outcome ID...")
	// Get the parameters
	outcomeID := request.QueryStringParameters[outcomeIDTag]
	// Configure the database connection
	db := setupDB()
	// Query the player data
	rows, err := db.Query("CALL GetOutcomeSection2(?);", outcomeID)
	// Check the query for issues
	if err != nil {
		log.Panicln(err)
	}
	// Store the returned data
	var data []ExtraInfo
	// Iterate through all of the rows
	for rows.Next() {
		// Store the next sequence
		result := ExtraInfo{}
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
