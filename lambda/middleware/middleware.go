package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt/v5"
)

func ValidateJWTMiddleware(next func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)) func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

		// extract headers
		tokenString := extractTokenFromHeaders(request.Headers)
		if tokenString == "" {
			return events.APIGatewayProxyResponse{
				Body:       "Missing Auth Token",
				StatusCode: http.StatusUnauthorized,
			}, nil
		}

		claims, err := parseToken(tokenString)

		if err != nil {
			return events.APIGatewayProxyResponse{
				Body:       "User Unauthorized",
				StatusCode: http.StatusUnauthorized,
			}, err
		}

		expires := int64(claims["expires"].(float64))

		if time.Now().Unix() > expires {
			return events.APIGatewayProxyResponse{
				Body:       "Token Expired",
				StatusCode: http.StatusUnauthorized,
			}, nil
		}

		return next(request)
	}
}

func extractTokenFromHeaders(headers map[string]string) string {
	authHeader, ok := headers["Authorization"]
	if !ok {
		return ""
	}
	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		return ""
	}
	return splitToken[1]
}

func parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		secret := []byte(os.Getenv("JWT_SECRET"))
		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid - unauthorized")
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return nil, fmt.Errorf("claims of unauthorized type")
	}

	return claims, nil

}
