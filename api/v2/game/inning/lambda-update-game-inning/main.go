package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
)

type Result struct {
	InningID int `json:"history_id"`
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
	inningIDTag          = os.Getenv("INNING_ID_TAG")
	gameIDTag            = os.Getenv("GAME_ID_TAG")
	halfInningTag        = os.Getenv("HALF_INNING_TAG")
	homeTotalUnearnedTag = os.Getenv("HOME_TOTAL_UNEARNED_TAG")
	homeTotalEarnedTag   = os.Getenv("HOME_TOTAL_EARNED_TAG")
	awayTotalUnearnedTag = os.Getenv("AWAY_TOTAL_UNEARNED_TAG")
	awayTotalEarnedTag   = os.Getenv("AWAY_TOTAL_EARNED_TAG")
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
	inningID := request.QueryStringParameters[inningIDTag]
	gameID := request.QueryStringParameters[gameIDTag]
	halfInning := request.QueryStringParameters[halfInningTag]
	homeTotalEarned := request.QueryStringParameters[homeTotalEarnedTag]
	homeTotalUnearned := request.QueryStringParameters[homeTotalUnearnedTag]
	awayTotalEarned := request.QueryStringParameters[awayTotalEarnedTag]
	awayTotalUnearned := request.QueryStringParameters[awayTotalUnearnedTag]
	// Log the beginning of the request
	log.Println("Creating new game inning entry...")
	// Configure the database connection
	db := setupDB()
	// Query the ball flight data
	err := db.QueryRow(`CALL UpdateGameInning (?, ?, ?, ?, ?, ?, ?)`, inningID, gameID, halfInning, homeTotalUnearned, homeTotalEarned,
		awayTotalUnearned, awayTotalEarned)
	// Check the query for issues
	if err != nil {
		log.Panicln(err)
	}
	log.Println("Game inning with id: " + string(inningID) + " has been updated successfully")
	// Return the result
	return events.APIGatewayProxyResponse{Body: "Updated Successfully", StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
