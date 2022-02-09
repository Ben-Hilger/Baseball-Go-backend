package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Info struct {
	ZoneX int     `json:"zone_x"`
	ZoneY int     `json:"zone_y"`
	Val   float64 `json:"value"`
}

var (
	countedData        map[string]map[string]interface{} = make(map[string]map[string]interface{})
	paSet                                                = make(map[string]map[int64]bool)
	bipTag             string                            = "bip"
	strikeSwingingTag  string                            = "strikes_swinging"
	foulBallTag        string                            = "foul_balls"
	totalPitchesTag    string                            = "pitches"
	totalSwingsTag     string                            = "total_swings"
	totalBallsTag      string                            = "total_balls"
	hbpTag             string                            = "total_hbp"
	numSinglesTag      string                            = "num_singles"
	numDoublesTag      string                            = "num_doubles"
	numTriplesTag      string                            = "num_triples"
	numHRTag           string                            = "num_homeruns"
	numSacFlyTag       string                            = "num_sac_fly"
	numSacBuntTag      string                            = "num_sac_bunt"
	numPATag           string                            = "num_pa"
	numWalksTag        string                            = "num_walks"
	nummStrikesOutsTag string                            = "num_strike_outs"

	filterTag string = "filter"
	typeTag   string = "type"
	teamIDTag string = "teamID"
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	countedData = make(map[string]map[string]interface{})
	paSet = make(map[string]map[int64]bool)
	var response map[string]interface{} = make(map[string]interface{})
	teamID := request.QueryStringParameters[teamIDTag]
	filter := request.QueryStringParameters[filterTag]
	// numerator := request.QueryStringParameters[numeratorTag]
	// denominator := request.QueryStringParameters[denominatorTag]
	filterType := request.QueryStringParameters[typeTag]
	// numerFrac := cleanFracPart(numerator)
	// denomfrac := cleanFracPart(denominator)
	CountStats(filterType == "zone", filterType == "pitches", teamID, filter)
	response["data"] = countedData
	// var numResult map[string]map[int]map[int]int64
	// if len(numerFrac) != 0 {
	// 	numResult = processFracPart(numerFrac)
	// }
	// var denomResult map[string]map[int]map[int]int64
	// if len(denomfrac) != 0 {
	// 	denomResult = processFracPart(denomfrac)
	// }
	// fmt.Println(denomResult)
	// if len(numResult) > 0 && numerator != "" {
	// 	response["result"] = processFunction(numResult, denomResult, zoneFilter, denominator)
	// }
	res, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
	}
	return events.APIGatewayProxyResponse{Body: string(res), StatusCode: 200}, nil
}

func processFunction(numResult map[string]map[int]map[int]int64, denomResult map[string]map[int]map[int]int64, zoneFilter string, denominator string) map[string]interface{} {
	var formulaResult map[string]interface{} = make(map[string]interface{})
	for hitterID, zoneData := range numResult {
		var info []Info = make([]Info, 1)
		var num float64 = 0
		var denom float64 = 0
		for x, XData := range zoneData {
			for y := range XData {
				var numVal = float64(numResult[hitterID][x][y])
				var denomVal float64
				if _, ok := denomResult[hitterID]; ok {
					denomVal = float64(denomResult[hitterID][x][y])
				}
				if zoneFilter == "" {
					num += numVal
					denom += denomVal
				} else if zoneFilter == "all" {
					var val = numVal
					if denomVal != 0 {
						val /= denomVal
						fmt.Println("I'm  here", val)
					} else if denomVal == 0 && denominator != "" {
						val = 0
					}
					info = append(info, Info{
						ZoneX: x,
						ZoneY: y,
						Val:   val,
					})
				}
			}
		}
		if zoneFilter == "" {
			var val = num
			if denom != 0 {
				val /= denom
			} else if denom == 0 && denominator != "" {
				val = 0
			}
			formulaResult["all"] = val
		} else {
			formulaResult[hitterID] = info
		}
	}
	return formulaResult
}

func main() {

}

func isOperation(str string) bool {
	return false
}

func processFracPart(frac []string) map[string]map[int]map[int]int64 {
	fmt.Println(frac)
	var results map[string]map[int]map[int]int64 = make(map[string]map[int]map[int]int64)
	var currentOp string = "+"
	for _, val := range frac {
		if isOperation(val) {
			currentOp = val
		} else {
			for hitterID, value := range countedData {
				if _, ok := results[hitterID]; !ok {
					results[hitterID] = makeInitialMap()
				}
				if statData, ok := value[val]; ok {
					for zoneX, zoneXData := range statData.(map[int]map[int]int64) {
						for zoneY, zoneYData := range zoneXData {
							if currentOp == "+" {
								results[hitterID][zoneX][zoneY] += zoneYData
							} else if currentOp == "-" {
								results[hitterID][zoneX][zoneY] -= zoneYData
							} else if currentOp == "*" {
								results[hitterID][zoneX][zoneY] *= zoneYData
							}
						}
					}
				}
			}
		}
	}
	return results
}

func CountStats(isZone bool, isPitch bool, teamID string, filter string) {
	opt := option.WithCredentialsFile("key.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Panicln(err)
	}
	client, err := app.Firestore(context.Background())
	if err != nil {
		log.Panicln(err)
	}
	// Get all of the game data
	it := client.CollectionGroup("GameData").Documents(context.Background())
	var i int64 = 0
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}
		i += 1
		if err != nil {
			log.Fatalln(err)
		}
		data := doc.Data()
		var actions []interface{} = data["actions"].([]interface{})
		var hitter = data["hitter"].(map[string]interface{})
		var hitterID = hitter["first_name"].(string) + " " + hitter["last_name"].(string)
		if _, ok := countedData[hitterID]; !ok {
			countedData[hitterID] = makeInitialStatList(isZone, isPitch)
		}
		var gameID string = data["game_id"].(string)
		if _, ok := paSet[gameID]; !ok {
			paSet[gameID] = make(map[int64]bool)
		}
		for _, element := range actions {
			// Check if it's a pitch action
			if isPitchAction(element.(map[string]interface{})) {
				if isZone && (filter == "all") {
					processZoneEvent(hitterID, element.(map[string]interface{}), actions, data["state_of_game_before"].(map[string]interface{}), data["pa_number"].(int64), gameID)
				} else if isPitch && (filter == "all") {
					processPitchEvent(hitterID, element.(map[string]interface{}), actions, data["state_of_game_before"].(map[string]interface{}), data["pa_number"].(int64), gameID)
				} else {
					processTotalEvent(hitterID, element.(map[string]interface{}), actions, data["state_of_game_before"].(map[string]interface{}), data["pa_number"].(int64), gameID)
				}
			}
		}
	}
	fmt.Println(i)
}

func makeInitialStatList(isZone bool, isPitch bool) map[string]interface{} {
	if isZone {
		return map[string]interface{}{
			bipTag:             makeInitialMap(),
			strikeSwingingTag:  makeInitialMap(),
			foulBallTag:        makeInitialMap(),
			totalPitchesTag:    makeInitialMap(),
			totalSwingsTag:     makeInitialMap(),
			totalBallsTag:      makeInitialMap(),
			hbpTag:             makeInitialMap(),
			numSinglesTag:      makeInitialMap(),
			numDoublesTag:      makeInitialMap(),
			numTriplesTag:      makeInitialMap(),
			numHRTag:           makeInitialMap(),
			numWalksTag:        makeInitialMap(),
			numPATag:           makeInitialMap(),
			numWalksTag:        makeInitialMap(),
			nummStrikesOutsTag: makeInitialMap(),
			numSacFlyTag:       makeInitialMap(),
			numSacBuntTag:      makeInitialMap(),
		}
	} else if isPitch {
		return map[string]interface{}{
			bipTag:             makeInitialPitchMap(),
			strikeSwingingTag:  makeInitialPitchMap(),
			foulBallTag:        makeInitialPitchMap(),
			totalPitchesTag:    makeInitialPitchMap(),
			totalSwingsTag:     makeInitialPitchMap(),
			totalBallsTag:      makeInitialPitchMap(),
			hbpTag:             makeInitialPitchMap(),
			numSinglesTag:      makeInitialPitchMap(),
			numDoublesTag:      makeInitialPitchMap(),
			numTriplesTag:      makeInitialPitchMap(),
			numHRTag:           makeInitialPitchMap(),
			numWalksTag:        makeInitialPitchMap(),
			numPATag:           makeInitialPitchMap(),
			numWalksTag:        makeInitialPitchMap(),
			nummStrikesOutsTag: makeInitialPitchMap(),
			numSacFlyTag:       makeInitialPitchMap(),
			numSacBuntTag:      makeInitialPitchMap(),
		}
	} else {
		return map[string]interface{}{
			bipTag:             0,
			strikeSwingingTag:  0,
			foulBallTag:        0,
			totalPitchesTag:    0,
			totalSwingsTag:     0,
			totalBallsTag:      0,
			hbpTag:             0,
			numSinglesTag:      0,
			numDoublesTag:      0,
			numTriplesTag:      0,
			numHRTag:           0,
			numWalksTag:        0,
			numPATag:           0,
			numWalksTag:        0,
			nummStrikesOutsTag: 0,
			numSacFlyTag:       0,
			numSacBuntTag:      0,
		}
	}
}

func makeInitialPitchMap() map[string]int64 {
	return map[string]int64{
		"FB": 0,
		"SL": 0,
		"CH": 0,
		"CU": 0,
		"SP": 0,
		"CT": 0,
	}
}

func makeInitialMap() map[int]map[int]int64 {
	return map[int]map[int]int64{
		1: {
			1: 0,
			2: 0,
			3: 0,
			4: 0,
			5: 0,
		},
		2: {
			1: 0,
			2: 0,
			3: 0,
			4: 0,
			5: 0,
		},
		3: {
			1: 0,
			2: 0,
			3: 0,
			4: 0,
			5: 0,
		},
		4: {
			1: 0,
			2: 0,
			3: 0,
			4: 0,
			5: 0,
		},
		5: {
			1: 0,
			2: 0,
			3: 0,
			4: 0,
			5: 0,
		},
	}
}

func processTotalEvent(hitterID string, action map[string]interface{}, actions []interface{}, state map[string]interface{}, paNumber int64, gameID string) {
	// Total Swings
	if didSwing(action) {
		if _, ok := countedData[hitterID][totalSwingsTag]; !ok {
			countedData[hitterID][totalSwingsTag] = 0
		}
		countedData[hitterID][totalSwingsTag] = 1 + countedData[hitterID][totalSwingsTag].(int)
	}
	// Total Strikes Swinging
	if didSwing(action) && isStrike(action) {
		if _, ok := countedData[hitterID][strikeSwingingTag]; !ok {
			countedData[hitterID][strikeSwingingTag] = 0
		}
		countedData[hitterID][strikeSwingingTag] = 1 + countedData[hitterID][strikeSwingingTag].(int)
	}
	// Foul Balls
	if isFoulBall(action) {
		if _, ok := countedData[hitterID][foulBallTag]; !ok {
			countedData[hitterID][foulBallTag] = 0
		}
		countedData[hitterID][foulBallTag] = 1 + countedData[hitterID][foulBallTag].(int)
	}
	// Balls
	if isBall(action) {
		if _, ok := countedData[hitterID][totalBallsTag]; !ok {
			countedData[hitterID][totalBallsTag] = 0
		}
		countedData[hitterID][totalBallsTag] = 1 + countedData[hitterID][totalBallsTag].(int)
	}
	// Ball in Play
	if isBIP(action) {
		if _, ok := countedData[hitterID][bipTag]; !ok {
			countedData[hitterID][bipTag] = 0
		}
		countedData[hitterID][bipTag] = 1 + countedData[hitterID][bipTag].(int)
	}
	// Hit By Pitch
	if isHBP(action) {
		if _, ok := countedData[hitterID][hbpTag]; !ok {
			countedData[hitterID][hbpTag] = 0
		}
		countedData[hitterID][hbpTag] = 1 + countedData[hitterID][hbpTag].(int)
	}
	// Total Singles
	if isSingle(action) {
		if _, ok := countedData[hitterID][numSinglesTag]; !ok {
			countedData[hitterID][numSinglesTag] = 0
		}
		countedData[hitterID][numSinglesTag] = 1 + countedData[hitterID][numSinglesTag].(int)
	}
	// Total Doubles
	if isDouble(action) {
		if _, ok := countedData[hitterID][numDoublesTag]; !ok {
			countedData[hitterID][numDoublesTag] = 0
		}
		countedData[hitterID][numDoublesTag] = 1 + countedData[hitterID][numDoublesTag].(int)
	}
	// Total Triples
	if isTriple(action) {
		if _, ok := countedData[hitterID][numTriplesTag]; !ok {
			countedData[hitterID][numTriplesTag] = 0
		}
		countedData[hitterID][numTriplesTag] = 1 + countedData[hitterID][numTriplesTag].(int)
	}
	// Total HR
	if isHR(action) {
		if _, ok := countedData[hitterID][numHRTag]; !ok {
			countedData[hitterID][numHRTag] = 0
		}
		countedData[hitterID][numHRTag] = 1 + countedData[hitterID][numHRTag].(int)
	}
	// Count Pitch
	if _, ok := countedData[hitterID][totalPitchesTag]; !ok {
		countedData[hitterID][totalPitchesTag] = 0
	}
	// Count Walks
	if isWalk(action, state) {
		if _, ok := countedData[hitterID][numWalksTag]; !ok {
			countedData[hitterID][numWalksTag] = 0
		}
		countedData[hitterID][numWalksTag] = 1 + countedData[hitterID][numWalksTag].(int)
	}
	// Count Strike Outs
	if isStrikeOut(action, state) {
		if _, ok := countedData[hitterID][nummStrikesOutsTag]; !ok {
			countedData[hitterID][nummStrikesOutsTag] = 0
		}
		countedData[hitterID][nummStrikesOutsTag] = 1 + countedData[hitterID][nummStrikesOutsTag].(int)
	}
	// Count Sac Fly
	if isSacFly(action, actions) {
		if _, ok := countedData[hitterID][numSacFlyTag]; !ok {
			countedData[hitterID][numSacFlyTag] = 0
		}
		countedData[hitterID][numSacFlyTag] = 1 + countedData[hitterID][numSacFlyTag].(int)
	}
	// Count Sac Bunt
	if isSacBunt(action, actions) {
		if _, ok := countedData[hitterID][numSacBuntTag]; !ok {
			countedData[hitterID][numSacBuntTag] = 0
		}
		countedData[hitterID][numSacBuntTag] = 1 + countedData[hitterID][numSacBuntTag].(int)
	}
	// Count PA
	var name = gameID + hitterID
	if array, ok := paSet[name]; ok {
		if _, ok := array[paNumber]; !ok {
			if _, ok := countedData[hitterID][numPATag]; !ok {
				countedData[hitterID][numPATag] = 0
			}
			countedData[hitterID][numPATag] = 1 + countedData[hitterID][numPATag].(int)
			paSet[name][paNumber] = true
		}
	} else {
		paSet[name] = make(map[int64]bool)
		if _, ok := countedData[hitterID][numPATag]; !ok {
			countedData[hitterID][numPATag] = 0
		}
		countedData[hitterID][numPATag] = 1 + countedData[hitterID][numPATag].(int)
		paSet[name][paNumber] = true
	}
	countedData[hitterID][totalPitchesTag] = 1 + countedData[hitterID][totalPitchesTag].(int)
}

func processPitchEvent(hitterID string, action map[string]interface{}, actions []interface{}, state map[string]interface{}, paNumber int64, gameID string) {
	var pitch string
	if _, ok := action["pitch_thrown"].(map[string]interface{}); !ok {
		return
	} else {
		pitch = action["pitch_thrown"].(map[string]interface{})["short_name"].(string)
		if pitch == "FA" || pitch == "FT" {
			pitch = "FB"
		}
	}
	// Total Swings
	if didSwing(action) {
		countedData[hitterID][totalSwingsTag].(map[string]int64)[pitch] += 1
	}
	// Total Strikes Swinging
	if didSwing(action) && isStrike(action) {
		countedData[hitterID][strikeSwingingTag].(map[string]int64)[pitch] += 1
	}
	// Foul Balls
	if isFoulBall(action) {
		countedData[hitterID][foulBallTag].(map[string]int64)[pitch] += 1
	}
	// Balls
	if isBall(action) {
		countedData[hitterID][totalBallsTag].(map[string]int64)[pitch] += 1
	}
	// Ball in Play
	if isBIP(action) {
		countedData[hitterID][bipTag].(map[string]int64)[pitch] += 1
	}
	// Hit By Pitch
	if isHBP(action) {
		countedData[hitterID][hbpTag].(map[string]int64)[pitch] += 1
	}
	// Total Singles
	if isSingle(action) {
		countedData[hitterID][numSinglesTag].(map[string]int64)[pitch] += 1
	}
	// Total Doubles
	if isDouble(action) {
		countedData[hitterID][numDoublesTag].(map[string]int64)[pitch] += 1
	}
	// Total Triples
	if isTriple(action) {
		countedData[hitterID][numTriplesTag].(map[string]int64)[pitch] += 1
	}
	// Total HR
	if isHR(action) {
		countedData[hitterID][numHRTag].(map[string]int64)[pitch] += 1
	}
	// Count Walks
	if isWalk(action, state) {
		if _, ok := countedData[hitterID][numWalksTag]; !ok {
			countedData[hitterID][numWalksTag] = makeInitialMap()
		}
		countedData[hitterID][numWalksTag].(map[string]int64)[pitch] += 1
	}
	// Count Strike Outs
	if isStrikeOut(action, state) {
		if _, ok := countedData[hitterID][nummStrikesOutsTag]; !ok {
			countedData[hitterID][nummStrikesOutsTag] = makeInitialMap()
		}
		countedData[hitterID][nummStrikesOutsTag].(map[string]int64)[pitch] += 1
	}
	// Count Sac Fly
	if isSacFly(action, actions) {
		if _, ok := countedData[hitterID][numSacFlyTag]; !ok {
			countedData[hitterID][numSacFlyTag] = makeInitialMap()
		}
		countedData[hitterID][numSacFlyTag].(map[string]int64)[pitch] += 1
	}
	// Count Sac Bunt
	if isSacBunt(action, actions) {
		if _, ok := countedData[hitterID][numSacBuntTag]; !ok {
			countedData[hitterID][numSacBuntTag] = makeInitialMap()
		}
		countedData[hitterID][numSacBuntTag].(map[string]int64)[pitch] += 1
	}
	// Count PA
	var name = gameID + hitterID
	if array, ok := paSet[name]; ok {
		if _, ok := array[paNumber]; !ok {
			if _, ok := countedData[hitterID][numPATag]; !ok {
				countedData[hitterID][numPATag] = 0
			}
			countedData[hitterID][numPATag].(map[string]int64)[pitch] = 1 + countedData[hitterID][numPATag].(map[string]int64)[pitch]
			paSet[name][paNumber] = true
		}
	} else {
		paSet[name] = make(map[int64]bool)
		if _, ok := countedData[hitterID][numPATag]; !ok {
			countedData[hitterID][numPATag] = 0
		}
		countedData[hitterID][numPATag].(map[string]int64)[pitch] = 1 + countedData[hitterID][numPATag].(map[string]int64)[pitch]
		paSet[name][paNumber] = true
	}
	// Count Pitches
	countedData[hitterID][totalPitchesTag].(map[string]int64)[pitch] += 1
}

func processZoneEvent(hitterID string, action map[string]interface{}, actions []interface{}, state map[string]interface{}, paNumber int64, gameID string) {
	var zoneX int
	var zoneY int
	if _, ok := action["strike_X_rel"].(float64); !ok {
		return
	} else {
		zoneX = getZone(action["strike_X_rel"].(float64))
		zoneY = getZone(action["strike_Y_rel"].(float64))
	}
	// Total Swings
	if didSwing(action) {
		if _, ok := countedData[hitterID][totalSwingsTag]; !ok {
			countedData[hitterID][totalSwingsTag] = makeInitialMap()
		}
		countedData[hitterID][totalSwingsTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
	}
	// Total Strikes Swinging
	if didSwing(action) && isStrike(action) {
		if _, ok := countedData[hitterID][strikeSwingingTag]; !ok {
			countedData[hitterID][strikeSwingingTag] = makeInitialMap()
		}
		countedData[hitterID][strikeSwingingTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
	}
	// Foul Balls
	if isFoulBall(action) {
		if _, ok := countedData[hitterID][foulBallTag]; !ok {
			countedData[hitterID][foulBallTag] = makeInitialMap()
		}
		countedData[hitterID][foulBallTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
	}
	// Balls
	if isBall(action) {
		if _, ok := countedData[hitterID][totalBallsTag]; !ok {
			countedData[hitterID][totalBallsTag] = makeInitialMap()
		}
		countedData[hitterID][totalBallsTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
	}
	// Ball in Play
	if isBIP(action) {
		if _, ok := countedData[hitterID][bipTag]; !ok {
			countedData[hitterID][bipTag] = makeInitialMap()
		}
		countedData[hitterID][bipTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
	}
	// Hit By Pitch
	if isHBP(action) {
		if _, ok := countedData[hitterID][hbpTag]; !ok {
			countedData[hitterID][hbpTag] = makeInitialMap()
		}
		countedData[hitterID][hbpTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
	}
	// Total Singles
	if isSingle(action) {
		if _, ok := countedData[hitterID][numSinglesTag]; !ok {
			countedData[hitterID][numSinglesTag] = makeInitialMap()
		}
		countedData[hitterID][numSinglesTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
	}
	// Total Doubles
	if isDouble(action) {
		if _, ok := countedData[hitterID][numDoublesTag]; !ok {
			countedData[hitterID][numDoublesTag] = makeInitialMap()
		}
		countedData[hitterID][numDoublesTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
	}
	// Total Triples
	if isTriple(action) {
		if _, ok := countedData[hitterID][numTriplesTag]; !ok {
			countedData[hitterID][numTriplesTag] = makeInitialMap()
		}
		countedData[hitterID][numTriplesTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
	}
	// Total HR
	if isHR(action) {
		if _, ok := countedData[hitterID][numHRTag]; !ok {
			countedData[hitterID][numHRTag] = makeInitialMap()
		}
		countedData[hitterID][numHRTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
	}
	// Count Pitch
	if _, ok := countedData[hitterID][totalPitchesTag]; !ok {
		countedData[hitterID][totalPitchesTag] = makeInitialMap()
	}
	// Count Walks
	if isWalk(action, state) {
		if _, ok := countedData[hitterID][numWalksTag]; !ok {
			countedData[hitterID][numWalksTag] = makeInitialMap()
		}
		countedData[hitterID][numWalksTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
	}
	// Count Strike Outs
	if isStrikeOut(action, state) {
		if _, ok := countedData[hitterID][nummStrikesOutsTag]; !ok {
			countedData[hitterID][nummStrikesOutsTag] = makeInitialMap()
		}
		countedData[hitterID][nummStrikesOutsTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
	}
	// Count Sac Fly
	if isSacFly(action, actions) {
		if _, ok := countedData[hitterID][numSacFlyTag]; !ok {
			countedData[hitterID][numSacFlyTag] = makeInitialMap()
		}
		countedData[hitterID][numSacFlyTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
	}
	// Count Sac Bunt
	if isSacBunt(action, actions) {
		if _, ok := countedData[hitterID][numSacBuntTag]; !ok {
			countedData[hitterID][numSacBuntTag] = makeInitialMap()
		}
		countedData[hitterID][numSacBuntTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
	}
	// Count PA
	var name = gameID + hitterID
	if array, ok := paSet[name]; ok {
		if _, ok := array[paNumber]; !ok {
			if _, ok := countedData[hitterID][numPATag]; !ok {
				countedData[hitterID][numPATag] = 0
			}
			countedData[hitterID][numPATag].(map[int]map[int]int64)[zoneX][zoneY] = 1 + countedData[hitterID][numPATag].(map[int]map[int]int64)[zoneX][zoneY]
			paSet[name][paNumber] = true
		}
	} else {
		paSet[name] = make(map[int64]bool)
		if _, ok := countedData[hitterID][numPATag]; !ok {
			countedData[hitterID][numPATag] = 0
		}
		countedData[hitterID][numPATag].(map[int]map[int]int64)[zoneX][zoneY] = 1 + countedData[hitterID][numPATag].(map[int]map[int]int64)[zoneX][zoneY]
		paSet[name][paNumber] = true
	}
	countedData[hitterID][totalPitchesTag].(map[int]map[int]int64)[zoneX][zoneY] += 1
}

func ZoneInFilter(zone int, filter string) bool {
	return strings.Contains(filter, string(zone))
}

func getZone(strikeX float64) int {
	return int(strikeX*5) + 1
}

func isPitchAction(action map[string]interface{}) bool {
	return action["action_type"].(int64) == 4
}

func didSwing(action map[string]interface{}) bool {
	// Get the action type
	if isPitchAction(action) {
		var outcomeInfo1 map[string]interface{} = action["extra_info_1"].(map[string]interface{})
		if outcomeInfo1["hitter_swing"].(bool) {
			return true
		}
		if outcomeInfo2, ok := action["extra_info_2"].(map[string]interface{}); ok {
			if outcomeInfo2["hitter_swing"].(bool) {
				return true
			}
		}

	}
	return false
}

func isBIP(action map[string]interface{}) bool {
	// Get the action type
	if isPitchAction(action) {
		var outcomeInfo map[string]interface{} = action["pitch_outcome"].(map[string]interface{})
		if val, ok := outcomeInfo["is_bip"].(bool); ok && val {
			return true
		}
	}
	return false
}

func isFoulBall(action map[string]interface{}) bool {
	// Get the action type
	if isPitchAction(action) {
		var outcomeInfo map[string]interface{} = action["pitch_outcome"].(map[string]interface{})
		if val, ok := outcomeInfo["is_fb"].(bool); ok && val {
			return true
		}
	}
	return false
}

func isHBP(action map[string]interface{}) bool {
	// Get the action type
	if isPitchAction(action) {
		var outcomeInfo map[string]interface{} = action["pitch_outcome"].(map[string]interface{})
		if val, ok := outcomeInfo["is_hbp"].(bool); ok && val {
			return true
		}
	}
	return false
}

func isStrike(action map[string]interface{}) bool {
	// Check that it is a pitch action
	if isPitchAction(action) {
		var outcomeInfo map[string]interface{} = action["pitch_outcome"].(map[string]interface{})
		if outcomeInfo["strikes_to_add"].(int64) > 0 && !outcomeInfo["is_fb"].(bool) {
			return true
		}
	}
	return false
}

func isBall(action map[string]interface{}) bool {
	// Check that it is a pitch action
	if isPitchAction(action) {
		var outcomeInfo map[string]interface{} = action["pitch_outcome"].(map[string]interface{})
		if outcomeInfo["balls_to_add"].(int64) > 0 {
			return true
		}
	}
	return false
}

func isSingle(action map[string]interface{}) bool {
	// Check that it is a pitch action
	if isPitchAction(action) && isBIP(action) {
		if extraInfo, ok := action["extra_info_2"]; ok {
			if extraInfo.(map[string]interface{})["send_hitter_to_base_id"].(int64) == 1 {
				return true
			}
		}
	}
	return false
}

func isDouble(action map[string]interface{}) bool {
	// Check that it is a pitch action
	if isPitchAction(action) && isBIP(action) {
		if extraInfo, ok := action["extra_info_2"]; ok {
			if extraInfo.(map[string]interface{})["send_hitter_to_base_id"].(int64) == 2 {
				return true
			}
		}
	}
	return false
}

func isTriple(action map[string]interface{}) bool {
	// Check that it is a pitch action
	if isPitchAction(action) && isBIP(action) {
		if extraInfo, ok := action["extra_info_2"]; ok {
			if extraInfo.(map[string]interface{})["send_hitter_to_base_id"].(int64) == 3 {
				return true
			}
		}
	}
	return false
}

func isHR(action map[string]interface{}) bool {
	// Check that it is a pitch action
	if isPitchAction(action) && isBIP(action) {
		if extraInfo, ok := action["extra_info_2"]; ok {
			if extraInfo.(map[string]interface{})["send_hitter_to_base_id"].(int64) == 4 {
				return true
			}
		}
	}
	return false
}

func runnersAdvanced(actions []interface{}) bool {
	for _, action := range actions {
		if action.(map[string]interface{})["action_type"].(int64) == 13 {
			return true
		}
	}
	return false
}

func isFlyBall(action map[string]interface{}) bool {
	info1 := action["extra_info_1"].(map[string]interface{})
	return info1["long_name"].(string) == "Fly Ball"
}

func isBunt(action map[string]interface{}) bool {
	info1 := action["extra_info_1"].(map[string]interface{})
	return info1["long_name"].(string) == "Bunt"
}

func hitterIsOut(action map[string]interface{}) bool {
	info1 := action["extra_info_1"].(map[string]interface{})
	if info1["mark_hitter_as_out"].(bool) {
		return true
	}
	if info2, ok := action["extra_info_2"].(map[string]interface{}); ok {
		if info2["mark_hitter_as_out"].(bool) {
			return true
		}
	}
	return false
}

func isSacFly(action map[string]interface{}, other []interface{}) bool {
	if isBIP(action) && isFlyBall(action) && runnersAdvanced(other) && hitterIsOut(action) {
		return true
	}
	return false
}

func isSacBunt(action map[string]interface{}, other []interface{}) bool {
	return isBIP(action) && runnersAdvanced(other) && hitterIsOut(action)
}

func isWalk(action map[string]interface{}, state map[string]interface{}) bool {
	return state["number_balls"].(int64)+action["pitch_outcome"].(map[string]interface{})["balls_to_add"].(int64) >= 4
}

func isStrikeOut(action map[string]interface{}, state map[string]interface{}) bool {
	if state["number_strikes"].(int64)+action["pitch_outcome"].(map[string]interface{})["strikes_to_add"].(int64) >= 3 && !action["pitch_outcome"].(map[string]interface{})["is_fb"].(bool) {
		return true
	}
	return false
}
