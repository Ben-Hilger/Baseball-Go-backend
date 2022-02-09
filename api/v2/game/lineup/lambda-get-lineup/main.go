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

type Result struct {
	Data []Lineup `json:"lineup_id"`
}

type Lineup struct {
	LineupMemberID int    `json:"lineup_member_id"`
	NumberInLineup int    `json:"number_in_lineup"`
	PositionID     int    `json:"position_id"`
	DHPersonID     string `json:"dh_person_id"`
	PersonID       string `json:"person_id"`
	FirstName      string `json:"person_first_name"`
	LastName       string `json:"person_last_name"`
	Bio            string `json:"person_bio"`
	Height         int    `json:"person_height"`
	HomeTown       string `json:"person_home_town"`
	Nickname       string `json:"person_nickname"`
	Weight         int    `json:"person_weight"`
	DHFirstName    string `json:"dh_first_name"`
	DHLastName     string `json:"dh_last_name"`
	DHBio          string `json:"dh_bio"`
	DHHeight       int    `json:"dh_height"`
	DHHomeTown     string `json:"dh_home_town"`
	DHNickname     string `json:"dh_nickname"`
	DHWeight       int    `json:"dh_weight"`
}

var (
	lineupIDTag = os.Getenv("LINEUP_ID_TAG")
	gameIDTag   = os.Getenv("GAME_ID_TAG")
	teamIDTag   = os.Getenv("TEAM_ID_TAG")
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
	// Log the beginning of the request
	log.Println("Getting Lineup information...")
	// Configure the database connection
	db := setupDB()
	// Query the ball flight data
	rows, err := db.Query(`CALL GetLineup (?, ?, ?)`, lineupID, gameID, teamID)
	// Check the query for issues
	if err != nil {
		log.Panicln(err)
	}
	// Store the returned data
	var data []Lineup
	// Iterate through all of the rows
	for rows.Next() {
		// Store the next sequence
		result := Lineup{}
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
	// Check the marshal for issues
	if err != nil {
		log.Panicln(err)
	}
	log.Println("Aquired Lineup Information successfully")
	// Return the result
	return events.APIGatewayProxyResponse{Body: string(res), StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
