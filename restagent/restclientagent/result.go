package restclientagent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent"
	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent/endpoints"
)

// Functions for making REST API calls to get the voting result:
// http://localhost:8080/result

func (rca *RestClientBallotAgent) treatResponseResults(r *http.Response) (resp restagent.ResponseResult, err error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	err = json.Unmarshal(buf.Bytes(), &resp)

	return
}

func (rca *RestClientBallotAgent) doRequestResults(ballotId string) (res restagent.ResponseResult, err error) {

	// Serialize the request
	url := rca.url + endpoints.Results

	// Create the request
	req := restagent.RequestResult{
		BallotId: ballotId,
	}

	// Send the request
	data, err := json.Marshal(req)
	if err != nil {
		return res, fmt.Errorf("/result. Error by %s in /result while marshalling request: %s", rca.Id, err.Error())
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))

	// Handle the response
	if err != nil {
		return res, fmt.Errorf("/result. Error by %s in /result while sending request: %s", rca.Id, err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("[%d] %s", resp.StatusCode, resp.Status)
		return
	}

	res, err = rca.treatResponseResults(resp)
	if err != nil {
		return res, fmt.Errorf("/result. Error by %s in /result while treating response: %s", rca.Id, err.Error())
	}

	return
}
