package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	httpPostURL := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=[API_KEY]"
	emailTag = "email"
	passwordTag = "password"
	secureTokenTag = "authID"
)

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get the secure token
	token = request.QueryStringParameters[secureTokenTag]
	if token != "" {

	}
	// Get the email and password
	email := request.QueryStringParameters[emailTag]
	password := request.QueryStringParameters[passwordTag]
	// Log the email
	fmt.Println("Making Authorization Post Call for email: " + email)
	var jsonData = []byte(`
		"email":` + email + `
		"password":` + password + `
		"returnSecureToken": true`)
	request, err := http.NewRequest("POST", httpPhttpPostURL, bytes.newBuffer(jsonData))
	if err != nil {
		log.Println(err)
	}
	defer response.Body.Close()
	if (response.Status )
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	// Return the result
	return events.APIGatewayProxyResponse{Body: "", StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
