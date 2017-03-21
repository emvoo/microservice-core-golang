package response

import (
	"log"
	"regexp"
	"strings"
)

// MicroserviceReponse - A standardised reponse format for a microservice.
type MicroserviceReponse struct {
	Status  string                 `json:"status"`
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// CreateResponse - Prepare a response for a microservice endpoint.
//
// This ensures that all API endpoints return data in a standardised format:
//
// {
//     "status": "ok", - Can contain any string. Usually 'ok', 'error' etc.
//     "code": 200, - A HTTP status code.
//     "message": "", - A message string elaborating on the status.
//     "data": {[ - A collection of return data. Can be omitted in the event
//     ]}           an error occurred.
// }
//
// Params:
//     Type string - The type of data being returned. Will be used to name the
//     collection.
//     Data interface{} - The data to return. Will always be parsed into a
//     collection.
//     Code int - HTTP status code for the response.
//     Status - A short status message. Examples: 'OK', 'Bad Request',
//     'Not Found'.
//     Message string - A more detailed status message.
//
// Return:
//     *MicroserviceReponse - The populated response object.
func CreateResponse(Type string, Data interface{}, Code int, Status string, Message string) *MicroserviceReponse {
	// Validate the arguments.
	if Data != nil && Type == "" {
		log.Fatal("Cannot prepare response. No type specified.")
	} else {
		reg, err := regexp.Compile("[^A-Za-z]")
		if err != nil {
			log.Fatal(err)
		}

		Type = reg.ReplaceAllString(strings.ToLower(Type), "-")
	}

	// Prepare the response object.
	response := &MicroserviceReponse{
		Status:  Status,
		Code:    Code,
		Message: Message,
	}

	// Only add the data if supplied.
	if Data != nil {
		response.Data = make(map[string]interface{})
		response.Data[Type] = Data
	}

	return response
}
