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

type TeamRoster struct {
	PersonID              string `json:"person_id"`
	FirstName             string `json:"first_name"`
	LastName              string `json:"last_name"`
	Bio                   string `json:"bio"`
	Height                int    `json:"height"`
	HomeTown              string `json:"home_town"`
	Nickname              string `json:"nickname"`
	Weight                int    `json:"nickname"`
	RoleShortName         string `json:"role_short_name"`
	RoleLongName          string `json:"role_long_name"`
	ThrowingHandShortName string `json:"throwing_hand_short_name"`
	ThrowingHandLongName  string `json:"throwing_hand_long_name"`
	HittingHandShortName  string `json:"hitting_hand_short_name"`
	HittingHandLongName   string `json:"hitting_hand_long_name"`
}

type Result struct {
	Data []TeamRoster `json:"data"`
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
	team_id_tag   = os.Getenv("TEAM_ID_TAG")
	season_id_tag = os.Getenv("SEASON_ID_TAG")
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get the teamID
	var teamID = request.QueryStringParameters[team_id_tag]
	// Get the seasonID
	var seasonID = request.QueryStringParameters[season_id_tag]
	log.Println("Getting team with id: " + teamID)
	// Configure the database connection
	db := setupDB()
	// Query the team data and scan the first row
	rows, err := db.Query("CALL GetTeamRoster (?, ?)", teamID, seasonID)
	// Check the query for issues
	if err != nil {
		log.Println(err)
	}
	// Store the returned data
	var data []TeamRoster
	// Iterate through all of the rows
	for rows.Next() {
		// Store the next sequence
		result := TeamRoster{}
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
	// Marshall the json result
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
