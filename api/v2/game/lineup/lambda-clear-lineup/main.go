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
	LineupID int `json:"lineup_id"`
}

var (
	gameIDTag   = os.Getenv("GAME_ID_TAG")
	teamIDTag   = os.Getenv("TEAM_ID_TAG")
	lineupIDTag = os.Getenv("LINEUP_ID_TAG")
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
	teamID := request.QueryStringParameters[teamIDTag]
	lineupID := request.QueryStringParameters[lineupIDTag]
	// Log the beginning of the request
	log.Println("Creating clearing lineup entry...")
	// Configure the database connection
	db := setupDB()
	// Query the ball flight data
	err := db.QueryRow(`CALL ClearLineup (?, ?, ?)`, lineupID, gameID, teamID)
	// Check the query for issues
	if err != nil {
		log.Panicln(err)
	}
	log.Println("Lineup with id: " + string(lineupID) + " has been cleared")
	// Return the result
	return events.APIGatewayProxyResponse{Body: "Cleared", StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
