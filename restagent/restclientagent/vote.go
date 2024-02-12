package restclientagent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent"
	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent/endpoints"
)

// Functions for making REST API calls to vote:
// http://localhost:8080/vote

func (rca *RestClientVoteAgent) doRequestVote(req restagent.RequestVote) (err error) {

	// Serialize the request
	url := rca.url + endpoints.Vote
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("/vote. Error by %s in /vote while marshalling request: %s", rca.Id, err.Error())
	}

	// Send the request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))

	// Handle the response
	if err != nil {
		return fmt.Errorf("/vote. Error by %s in /vote while sending request: %s", rca.Id, err.Error())
	}
	if resp.StatusCode != http.StatusOK {

		return fmt.Errorf("/vote. [%d] %s", resp.StatusCode, resp.Status)
	}
	return nil
}
