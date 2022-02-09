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
	GameID int `json:"game_id"`
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

var (
	gameIDTag          = os.Getenv("GAME_ID_TAG")
	awayTeamIDTag      = os.Getenv("AWAY_TEAM_ID_TAG")
	homeTeamIDTag      = os.Getenv("HOME_TEAM_ID_TAG")
	gameDateTag        = os.Getenv("GAME_DATE_ID_TAG")
	gameStartHourTag   = os.Getenv("GAME_START_HOUR_ID_TAG")
	gameStartMinuteTag = os.Getenv("GAME_START_MINUTE_ID_TAG")
	seasonIDTag        = os.Getenv("SEASON_ID_TAG")
	homeLineupIDTag    = os.Getenv("HOME_LINEUP_ID_TAG")
	awayLineupIDTag    = os.Getenv("AWAY_LINEUP_ID_TAG")
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get the parameters
	gameID := request.QueryStringParameters[gameIDTag]
	awayTeamID := request.QueryStringParameters[awayTeamIDTag]
	homeTeamID := request.QueryStringParameters[homeTeamIDTag]
	gameDate := request.QueryStringParameters[gameDateTag]
	gameStartHour := request.QueryStringParameters[gameStartHourTag]
	gameStartMinute := request.QueryStringParameters[gameStartMinuteTag]
	seasonID := request.QueryStringParameters[seasonIDTag]
	homeLineupID := request.QueryStringParameters[homeLineupIDTag]
	awayLineupID := request.QueryStringParameters[awayLineupIDTag]
	// Log the beginning of the request
	log.Println("Updating game...")
	// Configure the database connection
	db := setupDB()
	// Query the ball flight data
	db.QueryRow("CALL UpdateGame(?, ?, ?, ?, ?, ?, ?, ?, ?)", gameID, homeTeamID, awayTeamID,
		gameDate, gameStartHour, gameStartMinute, seasonID, homeLineupID, awayLineupID)
	// Return the result
	return events.APIGatewayProxyResponse{Body: "Updated Successfully", StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
