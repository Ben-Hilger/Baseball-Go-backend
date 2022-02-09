package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"

	firebase "firebase.google.com/go"

	"google.golang.org/api/option"
)

type DateType time.Time

type BallFlightResult struct {
	PersonID         string    `json:"person_id"`
	Time             time.Time `json:"time"`
	Velocity         float64   `json:"velocity"`
	Total_spin       int       `json:"total_spin"`
	True_spin        int       `json:"true_spin"`
	Spin_eff         float64   `json:"spin_eff"`
	Spin_dir_hour    int       `json:"spin_dir_hour"`
	Spin_dir_minute  int       `json:"spin_dir_minute"`
	Horizontal_break float64   `json:"horizontal_break"`
	Vertical_break   float64   `json:"vertical_break"`
	Is_strike        bool      `json:"is_strike"`
	Release_height   float64   `json:"release_height"`
	Release_side     float64   `json:"release_side"`
	Release_angle    float64   `json:"release_angle"`
	Horizontal_angle float64   `json:"horizontal_angle"`
	Gyro             float64   `json:"gyro"`
	Pitch_type       string    `json:"pitch_type"`
}

type Result struct {
	Data string `json:"data"`
}

func extractParameters(request events.APIGatewayProxyRequest) (float64, int, int, float64,
	int, int, float64, float64, bool, float64, float64, float64, float64, float64, string) {
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
	pitch_type := request.QueryStringParameters["pitch_type"]
	if pitch_type == "" {
		log.Println("No pitch type was given...")
	}
	// Return all of the extracted parameters
	return velocity, int(total_spin), int(true_spin), spin_eff, int(spin_dir_hour),
		int(spin_dir_minute), horizontal_break, vertical_break, is_strike, release_height,
		release_side, release_angle, horizontal_angle, gyro, pitch_type
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println(request.QueryStringParameters)
	// Get the user ID
	id := request.QueryStringParameters["personID"]
	// Extract the parameters
	velocity, total_spin, true_spin, spin_eff, spin_dir_hour, spin_dir_minute,
		horizontal_break, vertical_break, is_strike, release_height, release_side, release_angle,
		horizontal_angle, gyro, pitch_type := extractParameters(request)
	// Get the data
	date, err := time.Parse("01/02/2006", request.QueryStringParameters["time"])
	if err != nil {
		log.Fatal("There was an error reading the time: ", err)
	}
	// Get the required parameters
	ballFlightRequest := BallFlightResult{
		PersonID:         request.QueryStringParameters["personID"],
		Time:             date,
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
	// Configure Firebase
	opt := option.WithCredentialsFile("key.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "There was an issue processing the request", StatusCode: 500}, fmt.Errorf("error initializing app: %v", err)
	}
	client, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()
	_, _, issue := client.Collection("Data").Doc("Rapsodo").Collection(id).Add(context.Background(), ballFlightRequest)
	if issue != nil {
		log.Fatalf("Failed adding new data: ", issue)
	}
	return events.APIGatewayProxyResponse{Body: "Added successfully", StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
