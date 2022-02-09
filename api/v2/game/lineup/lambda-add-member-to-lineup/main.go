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
	LineupID int `json:"lineup_id"`
}

var (
	lineupIDTag       = os.Getenv("LINEUP_ID_TAG")
	gameIDTag         = os.Getenv("GAME_ID_TAG")
	teamIDTag         = os.Getenv("TEAM_ID_TAG")
	personIDTag       = os.Getenv("PERSON_ID_TAG")
	numberInLineupTag = os.Getenv("NUM_IN_LINEUP_TAG")
	positionIDTag     = os.Getenv("POSITION_ID_TAG")
	dhPersonIDTag     = os.Getenv("DH_PERSON_ID_TAG")
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
	personID := request.QueryStringParameters[personIDTag]
	numberInLineup := request.QueryStringParameters[numberInLineupTag]
	positionID := request.QueryStringParameters[positionIDTag]
	dhPersonID := request.QueryStringParameters[dhPersonIDTag]
	// Log the beginning of the request
	log.Println("Creating adding new member to lineup...")
	// Configure the database connection
	db := setupDB()
	var data int
	// Call the AddMemberToLineup function
	err := db.QueryRow(`CALL AddMemberToLineup (?, ?, ?, ?, ?, ?, ?)`, lineupID, gameID, teamID,
		personID, numberInLineup, positionID, dhPersonID).Scan(&data)
	// Check the query for issues
	if err != nil {
		log.Panicln(err)
	}
	res, err := json.Marshal(Result{LineupID: data})
	// Check the marshal for issues
	if err != nil {
		log.Panicln(err)
	}
	log.Println("New person added to lineup wth id: " + string(data) + " has been created")
	// Return the result
	return events.APIGatewayProxyResponse{Body: string(res), StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
