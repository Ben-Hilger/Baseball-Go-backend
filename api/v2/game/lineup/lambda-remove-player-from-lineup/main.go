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

var (
	lineupIDTag       = os.Getenv("LINEUP_ID_TAG")
	gameIDTag         = os.Getenv("GAME_ID_TAG")
	teamIDTag         = os.Getenv("TEAM_ID_TAG")
	lineupMemberIDTag = os.Getenv("LINEUP_MEMBER_ID_TAG")
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
	lineupID := request.QueryStringParameters[lineupIDTag]
	gameID := request.QueryStringParameters[gameIDTag]
	teamID := request.QueryStringParameters[teamIDTag]
	lineupMemberID := request.QueryStringParameters[lineupMemberIDTag]
	// Log the beginning of the request
	log.Println("Removing player from lineup...")
	// Configure the database connection
	db := setupDB()
	// Call the AddMemberToLineup function
	err := db.QueryRow(`CALL RemovePlayerFromLineup (?, ?, ?, ?)`, lineupID, gameID, teamID, lineupMemberID)
	if err != nil {
		log.Panicln(err)
	}
	log.Println("Removed player from lineup successfully")
	// Return the result
	return events.APIGatewayProxyResponse{Body: "Removed Successfully", StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
