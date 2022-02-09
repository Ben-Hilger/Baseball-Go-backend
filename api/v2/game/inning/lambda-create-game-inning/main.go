package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
)

type Result struct {
	InningID int `json:"inning_id"`
}

type Inning struct {
	GameID            int `json:"game_id"`
	HalfInning        int `json:"half_inning"`
	HomeTotalUnearned int `json:"home_total_unearned"`
	HomeTotalEarned   int `json:"home_total_earned"`
	AwayTotalUnearned int `json:"away_total_unearned"`
	AwayTotalEarned   int `json:"away_total_earned"`
}

var (
	gameIDTag     = os.Getenv("GAME_ID_TAG")
	halfInningTag = os.Getenv("HALF_INNING_TAG")
)

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
	// Get the parameters
	gameID := request.QueryStringParameters[gameIDTag]
	halfInning := request.QueryStringParameters[halfInningTag]
	// Log the beginning of the request
	log.Println("Creating new game inning entry...")
	// Configure the database connection
	db := setupDB()
	var data int
	// Query the ball flight data
	err := db.QueryRow(`CALL CreateGameInning (?, ?)`, gameID, halfInning).Scan(&data)
	// Check the query for issues
	if err != nil {
		log.Panicln(err)
	}
	res, err := json.Marshal(Result{InningID: data})
	// Check the marshal for issues
	if err != nil {
		log.Panicln(err)
	}
	log.Println("New game inning with id: " + string(data) + " has been created")
	// Return the result
	return events.APIGatewayProxyResponse{Body: string(res), StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
