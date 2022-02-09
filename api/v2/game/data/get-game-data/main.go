package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"

	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GameData struct {
	Actions       interface{} `json:"actions"`
	Hitter        interface{} `json:"hitter"`
	HitterHand    string      `json:"hitter_hand"`
	Pitcher       interface{} `json:"pitcher"`
	PitcherHand   string      `json:"pitcher_hand"`
	PitchingStyle string      `json:"pitching_style"`
	StateOfGame   interface{} `json:"state_of_game_before"`
}

func LoadGameData() []GameData {
	// Configure Firebase
	opt := option.WithCredentialsFile("key.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalln(err)
	}
	client, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()
	it := client.CollectionGroup("GameData").Documents(context.Background())
	var results []GameData
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		data := doc.Data()
		fmt.Println(data)
		hitterHand := ""
		pitcherHand := ""
		pitcherStyle := ""
		if data["hitter_hand_id"].(int64) == 1 {
			hitterHand = "LHH"
		} else if data["hitter_hand_id"].(int64) == 2 {
			hitterHand = "RHH"
		}
		if data["pitcher_hand_id"].(int64) == 1 {
			pitcherHand = "LHP"
		} else if data["pitcher_hand_id"].(int64) == 2 {
			pitcherHand = "RHP"
		}
		if data["pitching_style_id"].(int64) == 1 {
			pitcherStyle = "Stretch"
		} else if data["pitching_style_id"].(int64) == 2 {
			pitcherStyle = "Windup"
		}
		vals := GameData{
			Actions:       data["actions"],
			Hitter:        data["hitter"],
			HitterHand:    hitterHand,
			Pitcher:       data["pitcher"],
			PitcherHand:   pitcherHand,
			PitchingStyle: pitcherStyle,
			StateOfGame:   data["state_of_game_before"],
		}
		// data["pa_num"] = doc.Ref.ID
		results = append(results, vals)
	}
	return results
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println(request.QueryStringParameters)
	res, err := json.Marshal(LoadGameData())
	if err != nil {
		log.Println(err)
	}
	return events.APIGatewayProxyResponse{Body: string(res), StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
