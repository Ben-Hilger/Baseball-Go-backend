package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
)

type DateType time.Time

type BallFlightResult struct {
	PersonID         string  `json:"person_id"`
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
	Pitch_type       int     `json:"pitch_type"`
}

type Result struct {
	Data string `json:"data"`
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

func extractParameters(request events.APIGatewayProxyRequest) (float64, int, int, float64,
	int, int, float64, float64, bool, float64, float64, float64, float64, float64, int) {
	// Get the velocity
	velocity, err := strconv.ParseFloat(request.QueryStringParameters["velocity"], 64)
	if err != nil {
		log.Println(err)
	}
	// Get the total spin
	total_spin, err := strconv.ParseInt(request.QueryStringParameters["total_spin"], 0, 0)
	if err != nil {
		log.Println(err)
	}
	// Get the true spin
	true_spin, err := strconv.ParseInt(request.QueryStringParameters["true_spin"], 0, 0)
	if err != nil {
		log.Println(err)
	}
	// Get the spin efficiency
	spin_eff, err := strconv.ParseFloat(request.QueryStringParameters["spin_eff"], 64)
	if err != nil {
		log.Println(err)
	}
	// Get the spin direction hour
	spin_dir_hour, err := strconv.ParseInt(request.QueryStringParameters["spin_dir_hour"], 0, 0)
	if err != nil {
		log.Println(err)
	}
	// Get the spin direction minute
	spin_dir_minute, err := strconv.ParseInt(request.QueryStringParameters["spin_dir_minute"], 0, 0)
	if err != nil {
		log.Println(err)
	}
	// Get the horizontal break
	horizontal_break, err := strconv.ParseFloat(request.QueryStringParameters["horizontal_break"], 64)
	if err != nil {
		log.Println(err)
	}
	// Get the vertical break
	vertical_break, err := strconv.ParseFloat(request.QueryStringParameters["vertical_break"], 64)
	if err != nil {
		log.Println(err)
	}
	// Get if it was a strike. Default to false
	is_strike, err := strconv.ParseBool(request.QueryStringParameters["is_strike"])
	if err != nil {
		is_strike = false
	}
	// Get the release height
	release_height, err := strconv.ParseFloat(request.QueryStringParameters["release_height"], 64)
	if err != nil {
		log.Println(err)
	}
	// Get the release side
	release_side, err := strconv.ParseFloat(request.QueryStringParameters["release_side"], 64)
	if err != nil {
		log.Println(err)
	}
	// Get the release angle
	release_angle, err := strconv.ParseFloat(request.QueryStringParameters["release_angle"], 64)
	if err != nil {
		log.Println(err)
	}
	// Get the horizontal angle
	horizontal_angle, err := strconv.ParseFloat(request.QueryStringParameters["horizontal_angle"], 64)
	if err != nil {
		log.Println(err)
	}
	// Get the gyro
	gyro, err := strconv.ParseFloat(request.QueryStringParameters["gyro"], 64)
	if err != nil {
		log.Println(err)
	}
	// Get the pitch type
	pitch_type, err := strconv.ParseInt(request.QueryStringParameters["pitch_type"], 0, 0)
	if err != nil {
		log.Println(err)
	}
	// Return all of the extracted parameters
	return velocity, int(total_spin), int(true_spin), spin_eff, int(spin_dir_hour),
		int(spin_dir_minute), horizontal_break, vertical_break, is_strike, release_height,
		release_side, release_angle, horizontal_angle, gyro, int(pitch_type)
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println(request.QueryStringParameters)
	// Extract the parameters
	velocity, total_spin, true_spin, spin_eff, spin_dir_hour, spin_dir_minute,
		horizontal_break, vertical_break, is_strike, release_height, release_side, release_angle,
		horizontal_angle, gyro, pitch_type := extractParameters(request)
	// Get the required parameters
	ballFlightRequest := BallFlightResult{
		PersonID:         request.QueryStringParameters["personID"],
		Time:             request.QueryStringParameters["time"],
		Velocity:         velocity,
		Total_spin:       total_spin,
		True_spin:        true_spin,
		Spin_eff:         spin_eff,
		Spin_dir_hour:    spin_dir_hour,
		Spin_dir_minute:  spin_dir_minute,
		Horizontal_break: horizontal_break,
		Vertical_break:   vertical_break,
		Is_strike:        is_strike,
		Release_height:   release_height,
		Release_side:     release_side,
		Release_angle:    release_angle,
		Horizontal_angle: horizontal_angle,
		Gyro:             gyro,
		Pitch_type:       pitch_type,
	}
	// Log the beginning of the request
	log.Println("Uploading BallFlight Data for player: " +
		ballFlightRequest.PersonID)
	if ballFlightRequest.PersonID == "" {
		log.Panicln("No person ID is given, aborting!!")
	}
	// Configure the database connection
	db := setupDB()
	// Query the ball flight data
	stmt, err := db.Prepare("CALL UploadPlayerBallFlight (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	// Check the query for issues
	if err != nil {
		log.Panicln(err)
	}
	execResult, err := stmt.Exec(ballFlightRequest.PersonID, ballFlightRequest.Time,
		ballFlightRequest.Velocity, ballFlightRequest.Total_spin, ballFlightRequest.True_spin,
		ballFlightRequest.Spin_eff, ballFlightRequest.Spin_dir_hour, ballFlightRequest.Spin_dir_minute,
		ballFlightRequest.Horizontal_break, ballFlightRequest.Vertical_break, ballFlightRequest.Is_strike,
		ballFlightRequest.Release_height, ballFlightRequest.Release_side, ballFlightRequest.Release_angle,
		ballFlightRequest.Horizontal_angle, ballFlightRequest.Gyro, ballFlightRequest.Pitch_type)
	if err != nil {
		log.Panicln(err)
	}
	rowsAffected, err := execResult.RowsAffected()
	if err != nil {
		log.Panicln(err)
	}
	log.Println(string(rowsAffected) + " modified with query for player_id: " + string(ballFlightRequest.PersonID))
	return events.APIGatewayProxyResponse{Body: "Added successfully", StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
