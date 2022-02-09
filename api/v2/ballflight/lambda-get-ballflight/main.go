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

type DateType time.Time

type BallFlightResult struct {
	PlayerID         string  `json:"person_id"`
	Time             string  `json:"time"`
	Velocity         float64 `json:"velocity"`
	Total_spin       int     `json:"total_spin"`
	True_spin        int     `json:"true_spin"`
	Spin_eff         float64 `json:"spin_eff"`
	Spin_dir_hour    int     `json:"spin_dir_hour"`
	Spin_dir_minute  int     `json:"spin_dir_minute"`
	Horizontal_break float64 `json:"horizontal_break"`
	Vertical_break   float64 `json:"vertical_break"`
	Is_strike        bool    `json:"is_strike"`
	Release_height   float64 `json:"release_height"`
	Release_side     float64 `json:"release_side"`
	Release_angle    float64 `json:"release_angle"`
	Horizontal_angle float64 `json:"horizontal_angle"`
	Gyro             float64 `json:"gyro"`
	PitchShortName   string  `json:"short_name"`
	PitchLongName    string  `json:"long_name"`
}

type Result struct {
	Data []BallFlightResult `json:"data"`
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

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	playerID := string(request.QueryStringParameters["personID"])
	// Log the beginning of the request
	log.Println("Getting BallFlight Data for player: " + playerID)
	// Configure the database connection
	db := setupDB()
	// Query the ball flight data
	rows, err := db.Query("CALL GetPlayerBallFlight(?)", playerID)
	// Check the query for issues
	if err != nil {
		log.Panicln(err)
	}
	// Store the returned data
	var data []BallFlightResult
	// Iterate through all of the rows
	for rows.Next() {
		// Store the next sequence
		result := BallFlightResult{}
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
