package restserveragent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent"
	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent/comsoc"
)

// Functions that handle the REST API call to vote:
// http://localhost:8080/vote

// Decode the request
func (*RestServerAgent) decodeVoteRequest(r *http.Request) (req restagent.RequestVote, err error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		fmt.Println("Error decoding /vote request: ", err)
	}
	return
}

func checkVoteAlts(vote []comsoc.Alternative, expected int) bool {
	// Check if the vote matches the alternatives proposed by the ballot

	// Note: the expected alternatives range from 1 to Ballot.Alts (inclusive)
	if len(vote) != expected {
		return false
	}
	list := make([]comsoc.Alternative, expected)
	copy(list, vote)
	sort.Slice(list, func(i, j int) bool { return list[i] < list[j] })
	for i := 0; i < expected-1; i++ {
		if list[i] != list[i+1]-1 {
			return false
		}
	}
	return true
}

func checkVote(ballotsList map[string]restagent.Ballot, deadline time.Time, req restagent.RequestVote) (err error) {
	// Check if the ballot exists
	_, found := ballotsList[req.BallotId]
	if !found {
		return fmt.Errorf("notexist")
	}
	// Check if the agent has already voted
	for _, v := range ballotsList[req.BallotId].HaveVoted {
		if v == req.AgentId {
			return fmt.Errorf("alreadyvoted")
		}
	}
	// Check if the agent is allowed to vote
	var canVote bool
	for _, v := range ballotsList[req.BallotId].VoterIds {
		if v == req.AgentId {
			canVote = true
			break
		}
	}
	if !canVote {
		return fmt.Errorf("notallowed")
	}

	// Check if the deadline has passed
	if deadline.Before(time.Now()) {
		return fmt.Errorf("alreadyfinished")
	}

	// Check if the provided alternatives for the vote are correct
	if !checkVoteAlts(req.Prefs, ballotsList[req.BallotId].Alts) {
		return fmt.Errorf("wrongalts")
	}

	// If the ballot is "approval", check if a coherent threshold is provided
	if ballotsList[req.BallotId].Rule == "approval" {
		if req.Options == nil || len(req.Options) != 1 || req.Options[0] < 0 || req.Options[0] > ballotsList[req.BallotId].Alts {
			return fmt.Errorf("wrongthreshold")
		}
	}
	return nil
}

func (rsa *RestServerAgent) doVote(w http.ResponseWriter, r *http.Request) {
	rsa.Lock()
	defer rsa.Unlock()
	// Check the request method
	if !rsa.checkMethod("POST", w, r) {
		return
	}

	// Decode the request
	req, err := rsa.decodeVoteRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) //400
		fmt.Fprint(w, err.Error())
		return
	}

	// Check if the vote is correct
	err = checkVote(rsa.ballotsList, rsa.ballotsList[req.BallotId].Deadline, req)
	if err != nil {
		switch err.Error() {
		case "notexist":
			w.WriteHeader(http.StatusBadRequest) //400
			msg := fmt.Sprintf("error /vote: ballot %s does not exist", req.BallotId)
			w.Write([]byte(msg))
			return
		case "alreadyvoted":
			w.WriteHeader(http.StatusForbidden) //403
			msg := fmt.Sprintf("error /vote: agent %s has already voted for ballot %s", req.AgentId, req.BallotId)
			w.Write([]byte(msg))
			return
		case "notallowed":
			w.WriteHeader(http.StatusUnauthorized) //401
			msg := fmt.Sprintf("error /vote: agent %s is not allowed to vote for ballot %s", req.AgentId, req.BallotId)
			w.Write([]byte(msg))
			return
		case "alreadyfinished":
			w.WriteHeader(http.StatusServiceUnavailable) //503
			msg := fmt.Sprintf("error /vote: ballot %s is already finished: %s", req.BallotId, rsa.ballotsList[req.BallotId].Deadline.String())
			w.Write([]byte(msg))
			return
		case "wrongalts":
			w.WriteHeader(http.StatusBadRequest) //400
			msg := fmt.Sprintf("error /vote: alternatives provided for ballot %s are not correct", req.BallotId)
			w.Write([]byte(msg))
			return
		case "wrongthreshold":
			w.WriteHeader(http.StatusBadRequest) //400
			msg := fmt.Sprintf("error /vote: threshold %d provided for ballot %s is not correct", req.Options, req.BallotId)
			w.Write([]byte(msg))
			return
		}
	}

	// Save the threshold if necessary
	if rsa.ballotsList[req.BallotId].Rule == restagent.Approval {
		_, found := rsa.ballotsList[req.BallotId].Thresholds[req.AgentId]
		if found {
			w.WriteHeader(http.StatusBadRequest) //400
			msg := fmt.Sprintf("error /vote: agent %s has already provided a threshold for ballot %s", req.AgentId, req.BallotId)
			w.Write([]byte(msg))
			return
		}
		rsa.ballotsList[req.BallotId].Thresholds[req.AgentId] = req.Options[0]
	}

	// Save the vote for the ballot
	rsa.ballotsMap[req.BallotId] = append(rsa.ballotsMap[req.BallotId], req.Prefs)

	// Record that the agent has voted
	for i := 0; i < len(rsa.ballotsList[req.BallotId].HaveVoted); i++ {
		if rsa.ballotsList[req.BallotId].HaveVoted[i] == "" {
			rsa.ballotsList[req.BallotId].HaveVoted[i] = req.AgentId
			break
		}
	}

	w.WriteHeader(http.StatusOK) //200
	msg := "/vote: vote registered"
	w.Write([]byte(msg))
}
