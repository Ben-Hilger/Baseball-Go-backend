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
	HistoryID int `json:"history_id"`
}

type GameHistory struct {
	GameID              int     `json:"game_id"`
	InningID            int     `json:"inning_id"`
	HitterID            string  `json:"hitter_id"`
	PitcherID           string  `json:"pitcher_id"`
	PersonAtFirst       string  `json:"person_at_first"`
	PersonAtSecond      string  `json:"person_at_second"`
	PersonAtThird       string  `json:"person_at_third"`
	NumberOuts          int     `json:"number_outs"`
	NumberStrikes       int     `json:"number_strikes"`
	NumberBalls         int     `json:"number_balls"`
	CurrentPA           int     `json:"current_pa"`
	PitchNumber         int     `json:"pitch_number"`
	PitchThrown         int     `json:"pitch_thrown"`
	PitchStrikeX        float64 `json:"pitch_strike_zone_location_rel_X"`
	PitchStrikeY        float64 `json:"pitch_strike_zone_location_rel_Y"`
	BallLocX            float64 `json:"pitch_ball_field_location_rel_X"`
	BallLocY            float64 `json:"pitch_ball_field_location_rel_Y"`
	PitcherThrowingHand int     `json:"pitcher_throwing_hand"`
	HitterHittingHand   int     `json:"hitter_hitting_hand"`
	PitchOutcome        int     `json:"pitch_outcome_id"`
	PitchExtraInfo1     int     `json:"pitch_extra_info_1_id"`
	PitchExtraInfo2     int     `json:"pitch_extra_info_2_id"`
	PitchExtraInfo3     int     `json:"pitch_extra_info_3_id"`
	EventTypeID         int     `json:"event_type_id"`
	PersonInvolved      string  `json:"person_involved_id"`
	PersonBaseAt        int     `json:"person_base_at_id"`
	PersonBaseGoing     int     `json:"person_base_going_id"`
	NumberErrors        int     `json:"nunber_errors"`
	PitchVelo           float64 `json:"pitch_velo"`
	PitchingStyleID     int     `json:"pitching_style_id"`
	BallExitVelocity    float64 `json:"ball_exit_velocity"`
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
	gameIDTag              = os.Getenv("GAME_ID_TAG")
	inningIDTag            = os.Getenv("INNING_ID_TAG")
	hitterIDTag            = os.Getenv("HITTING_ID_TAG")
	pitcherIDTag           = os.Getenv("PITCHER_ID_TAG")
	personAtFirstTag       = os.Getenv("PERSON_AT_FIRST_TAG")
	personAtSecondTag      = os.Getenv("PERSON_AT_SECOND_TAG")
	personAtThirdTag       = os.Getenv("PERSON_AT_THIRD_TAG")
	numberOutsTag          = os.Getenv("NUMBER_OUTS_TAG")
	numberStrikesTag       = os.Getenv("NUMBER_STRIKES_TAG")
	numberBallsTag         = os.Getenv("NUMBER_BALLS_TAG")
	currentPATag           = os.Getenv("CURRENT_PA_TAG")
	pitchNumberTag         = os.Getenv("PITCH_NUMBER_TAG")
	pitchThrownTag         = os.Getenv("PITCH_THROWN_TAG")
	pitchStrikeXTag        = os.Getenv("PITCH_STRIKE_X_TAG")
	pitchStrikeYTag        = os.Getenv("PITCH_STRIKE_Y_TAG")
	pitchBallFieldXTag     = os.Getenv("PITCH_BALL_FIELD_X_TAG")
	pitchBallFieldYTag     = os.Getenv("PITCH_BALL_FIELD_Y_TAG")
	pitcherThrowingHandTag = os.Getenv("PITCHER_THROWING_HAND_TAG")
	hitterHittingHandTag   = os.Getenv("HITTER_HITTING_HAND_TAG")
	pitchOutcomeIDTag      = os.Getenv("PITCH_OUTCOME_ID_TAG")
	pitchExtraInfo1Tag     = os.Getenv("PITCH_EXTRA_INFO_1_TAG")
	pitchExtraInfo2Tag     = os.Getenv("PITCH_EXTRA_INFO_2_TAG")
	pitchExtraInfo3Tag     = os.Getenv("PITCH_EXTRA_INFO_3_TAG")
	eventTypeIDTag         = os.Getenv("EVENT_TYPE_ID_TAG")
	personInvolvedIDTag    = os.Getenv("PERSON_INVOLVED_ID_TAG")
	personBaseAtTag        = os.Getenv("PERSON_BASE_AT_TAG")
	personBaseGoingToTag   = os.Getenv("PERSON_BASE_GOING_TO_TAG")
	numberErrorsTag        = os.Getenv("NUMBER_ERRORS_TAG")
	pitchVeloTag           = os.Getenv("PITCH_VELO_TAG")
	pitchStylingIDTag      = os.Getenv("PITCH_STYLE_ID_TAG")
	ballExitVelocityTag    = os.Getenv("BALL_EXIT_VELOCITY_TAG")
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get the parameters
	gameID := request.QueryStringParameters[gameIDTag]
	inningID := request.QueryStringParameters[inningIDTag]
	hitterID := request.QueryStringParameters[hitterIDTag]
	pitcherID := request.QueryStringParameters[pitcherIDTag]
	personAtFirst := request.QueryStringParameters[personAtFirstTag]
	personAtSecond := request.QueryStringParameters[personAtSecondTag]
	personAtThird := request.QueryStringParameters[personAtThirdTag]
	numberOuts := request.QueryStringParameters[numberOutsTag]
	numberStrikes := request.QueryStringParameters[numberStrikesTag]
	numberBalls := request.QueryStringParameters[numberBallsTag]
	currentPA := request.QueryStringParameters[currentPATag]
	pitchNumber := request.QueryStringParameters[pitchNumberTag]
	pitchThrown := request.QueryStringParameters[pitchThrownTag]
	pitchStrikeX := request.QueryStringParameters[pitchStrikeXTag]
	pitchStrikeY := request.QueryStringParameters[pitchStrikeYTag]
	pitchBallFieldX := request.QueryStringParameters[pitchBallFieldXTag]
	pitchBallFieldY := request.QueryStringParameters[pitchBallFieldYTag]
	pitcherThrowingHand := request.QueryStringParameters[pitcherThrowingHandTag]
	hitterHittingHand := request.QueryStringParameters[hitterHittingHandTag]
	pitchOutcomeID := request.QueryStringParameters[pitchOutcomeIDTag]
	pitchExtraInfo1 := request.QueryStringParameters[pitchExtraInfo1Tag]
	pitchExtraInfo2 := request.QueryStringParameters[pitchExtraInfo2Tag]
	pitchExtraInfo3 := request.QueryStringParameters[pitchExtraInfo3Tag]
	eventTypeID := request.QueryStringParameters[eventTypeIDTag]
	personInvolvedID := request.QueryStringParameters[personInvolvedIDTag]
	personBaseAt := request.QueryStringParameters[personBaseAtTag]
	personBaseGoingTo := request.QueryStringParameters[personBaseGoingToTag]
	numberErrors := request.QueryStringParameters[numberErrorsTag]
	pitchVelo := request.QueryStringParameters[pitchVeloTag]
	pitchStylingID := request.QueryStringParameters[pitchStylingIDTag]
	ballExitVelocity := request.QueryStringParameters[ballExitVelocityTag]
	// Log the beginning of the request
	log.Println("Creating new game history entry...")
	// Configure the database connection
	db := setupDB()
	var data int
	// Query the ball flight data
	err := db.QueryRow(`CALL CreateGameHistory (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, gameID,
		inningID, hitterID, pitcherID, personAtFirst, personAtSecond, personAtThird,
		numberOuts, numberStrikes, numberBalls, currentPA, pitchNumber, pitchThrown, pitchStrikeX,
		pitchStrikeY, pitchBallFieldX, pitchBallFieldY, pitcherThrowingHand, hitterHittingHand,
		pitchOutcomeID, pitchExtraInfo1, pitchExtraInfo2, pitchExtraInfo3, eventTypeID,
		personInvolvedID, personBaseAt, personBaseGoingTo, numberErrors, pitchVelo,
		pitchStylingID, ballExitVelocity).Scan(&data)
	// Check the query for issues
	if err != nil {
		log.Panicln(err)
	}
	res, err := json.Marshal(Result{HistoryID: data})
	// Check the marshal for issues
	if err != nil {
		log.Panicln(err)
	}
	log.Println("New game history with id: " + string(data) + " has been created")
	// Return the result
	return events.APIGatewayProxyResponse{Body: string(res), StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
