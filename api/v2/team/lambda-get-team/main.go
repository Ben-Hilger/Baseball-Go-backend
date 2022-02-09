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

type Team struct {
	TeamID                string `json:"team_id"`
	Name                  string `json:"team_name"`
	Abbreviation          string `json:"abbreviation"`
	Location              string `json:"location"`
	Nickname              string `json:"nickname"`
	TeamPrimaryColorHex   string `json:"team_primary_color_hex"`
	TeamSecondaryColorHex string `json:"team_secondary_color_hex"`
	TeamTypeShortName     string `json:"team_type_short"`
	TeamTypeLongName      string `json:"team_type_long"`
}

type Result struct {
	Data Team `json:"data"`
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
	team_id_tag = "teamID"
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get the teamID
	var teamID = request.QueryStringParameters[team_id_tag]
	log.Println("Getting team with id: " + teamID)
	// Configure the database connection
	db := setupDB()
	var team Team
	// Query the team data and scan the first row
	err := db.QueryRow("CALL GetTeam(?)", teamID).Scan(&team.TeamID, &team.Name, &team.Abbreviation,
		&team.Location, &team.Nickname, &team.TeamPrimaryColorHex, &team.TeamSecondaryColorHex,
		&team.TeamTypeShortName, &team.TeamTypeLongName)
	// Check the query for issues
	if err != nil {
		log.Println(err)
	}
	// Marshall the json result
	res, err := json.Marshal(Result{Data: team})
	if err != nil {
		log.Println(err)
	}
	// Return the result
	return events.APIGatewayProxyResponse{Body: string(res), StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
