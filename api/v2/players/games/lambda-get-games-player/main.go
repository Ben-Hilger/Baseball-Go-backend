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

type Game struct {
	GameID          int    `json:"game_id"`
	GameDate        string `json:"game_date"`
	GameStartHour   int    `json:"game_start_hour"`
	GameStartMinute int    `json:"game_start_minute"`
	SeasonName      string `json:"season_name"`
	SeasonYear      int    `json:"season_year"`
	HomeTeamName    string `json:"home_team_name"`
	AwayTeamName    string `json:"away_team_name"`
}

type Result struct {
	Data []Game `json:"data"`
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
	playerTag = os.Getenv("PLAYER_ID_TAG")
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	playerID := request.QueryStringParameters[playerTag]
	log.Println("Getting Games for player: " + playerID)
	// Configure the database connection
	db := setupDB()
	// Query the ball flight data
	rows, err := db.Query("CALL GetPlayerGames(?)", playerID)
	// Check the query for issues
	if err != nil {
		log.Panicln(err)
	}
	// Store the returned data
	var data []Game
	// Iterate through all of the rows
	for rows.Next() {
		// Store the next sequence
		result := Game{}
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
	// Return the result
	return events.APIGatewayProxyResponse{Body: string(res), StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
