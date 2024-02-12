package restserveragent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent"
	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent/comsoc"
)

// Functions that handle the call to the REST API to get the vote result:
// http://localhost:8080/result

// Decode the request
func (*RestServerAgent) decodeResultRequest(r *http.Request) (req restagent.RequestResult, err error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		fmt.Println("Error decoding request /result: ", err)
		return
	}
	return
}

// Check the consistency of the request
func checkResultRequest(ballotsList map[string]restagent.Ballot, req restagent.RequestResult) (err error) {
	// Check if the ballot exists
	_, found := ballotsList[req.BallotId]
	if !found {
		return fmt.Errorf("notexist")
	}
	// Check if the deadline has passed
	if ballotsList[req.BallotId].Deadline.After(time.Now()) {
		return fmt.Errorf("notfinished")
	}

	// Check the consistency of thresholds (already checked upon receiving the vote request)
	// Note: possibly gaining in security but losing in performance
	if ballotsList[req.BallotId].Rule == restagent.Approval {
		var nbVoters int
		for ; nbVoters < len(ballotsList[req.BallotId].HaveVoted) && ballotsList[req.BallotId].HaveVoted[nbVoters] != ""; nbVoters++ {
		}

		if len(ballotsList[req.BallotId].Thresholds) != nbVoters {
			return fmt.Errorf("thresholdnumber")
		}
		for _, t := range ballotsList[req.BallotId].Thresholds {
			if t < 0 || t > ballotsList[req.BallotId].Alts {
				return fmt.Errorf("thresholdvalue")
			}
		}
	}
	return
}

// Calculate the vote result by applying the desired voting method
func (rsa *RestServerAgent) doCalcResult(w http.ResponseWriter, r *http.Request) {
	rsa.Lock()
	defer rsa.Unlock()
	// Check the request method
	if !rsa.checkMethod("POST", w, r) {
		return
	}

	req, err := rsa.decodeResultRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		fmt.Fprint(w, err.Error())
		return
	}

	// Check request
	err = checkResultRequest(rsa.ballotsList, req)
	if err != nil {
		switch err.Error() {
		case "notexist":
			w.WriteHeader(http.StatusNotFound) // 404
			msg := fmt.Sprintf("error /result: ballot %s does not exist", req.BallotId)
			w.Write([]byte(msg))
			return
		case "notfinished":
			w.WriteHeader(http.StatusTooEarly) // 425
			msg := fmt.Sprintf("error /result: ballot %s is not finished yet. Deadline: %s", req.BallotId, rsa.ballotsList[req.BallotId].Deadline)
			w.Write([]byte(msg))
			return
		case "thresholdnumber":
			w.WriteHeader(http.StatusBadRequest) // 400
			msg := fmt.Sprintf("error /result: ballot %s does not have the same number of thresholds and voters", req.BallotId)
			w.Write([]byte(msg))
			return
		case "thresholdvalue":
			w.WriteHeader(http.StatusBadRequest) // 400
			msg := fmt.Sprintf("error /result: ballot %s is approval and has a threshold value not in [0, %d]", req.BallotId, rsa.ballotsList[req.BallotId].Alts)
			w.Write([]byte(msg))
			return
		}
	}

	resp := restagent.ResponseResult{}

	// If no vote has been submitted, simply apply the tie-break (except for Condorcet where no Tie-Break is considered, returning 0)
	if len(rsa.ballotsMap[req.BallotId]) == 0 {
		// Note: we decide to return a result, but we could have returned an error
		if rsa.ballotsList[req.BallotId].Rule == restagent.Condorcet {
			resp.Winner = 0
		} else {
			resp.Winner = rsa.ballotsList[req.BallotId].TieBreak[0]
			resp.Ranking = rsa.ballotsList[req.BallotId].TieBreak
		}

		serial, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			msg := fmt.Sprintf("error /result: can't serialize response for ballot %s of type %s", req.BallotId, rsa.ballotsList[req.BallotId].Rule)
			w.Write([]byte(msg))
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		w.Write(serial)
		return
	}

	if rsa.ballotsList[req.BallotId].Rule == restagent.Approval {
		// Special case of Approval, as it requires an additional parameter (the threshold)

		// Transform the Threshold map into a list
		thresholds := make([]int, 0)
		for _, v := range rsa.ballotsList[req.BallotId].HaveVoted {
			if v == "" {
				break
			}
			thresholds = append(thresholds, rsa.ballotsList[req.BallotId].Thresholds[v])
		}
		tiebreak := comsoc.TieBreakFactory(rsa.ballotsList[req.BallotId].TieBreak)
		swf, err := comsoc.MakeApprovalRankingWithTieBreak(rsa.ballotsMap[req.BallotId], thresholds, tiebreak)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			msg := fmt.Sprintf("error /result: can't process SWF for ballot %s of type %s. "+err.Error(), req.BallotId, rsa.ballotsList[req.BallotId].Rule)
			w.Write([]byte(msg))
			return
		}

		resp.Winner = swf[0]
		resp.Ranking = swf

		serial, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			msg := fmt.Sprintf("error /result: can't serialize response for ballot %s of type %s", req.BallotId, rsa.ballotsList[req.BallotId].Rule)
			w.Write([]byte(msg))
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		w.Write(serial)
		return

	} else if rsa.ballotsList[req.BallotId].Rule == restagent.Condorcet {
		// Special case of Condorcet, as the calculation of SWF is not possible
		// Note: Tie-break is not used for Condorcet. Either a winner or none is returned
		scf, err := comsoc.CondorcetWinner(rsa.ballotsMap[req.BallotId])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			msg := fmt.Sprintf("error /result: can't process SCF for ballot %s of type %s", req.BallotId, rsa.ballotsList[req.BallotId].Rule)
			w.Write([]byte(msg))
			return
		}
		if len(scf) != 0 {
			resp.Winner = scf[0]
		}

		serial, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			msg := fmt.Sprintf("error /result: can't serialize response for ballot %s of type %s", req.BallotId, rsa.ballotsList[req.BallotId].Rule)
			w.Write([]byte(msg))
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		w.Write(serial)
		return
	} else if rsa.ballotsList[req.BallotId].Rule == restagent.STV {
		// Special case of STV, as the tie-break is not applied in the same way
		swf, err := comsoc.STV_SWF_TieBreak(rsa.ballotsMap[req.BallotId], rsa.ballotsList[req.BallotId].TieBreak)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			msg := fmt.Sprintf("error /result: can't process SWF for ballot %s of type %s. "+err.Error(), req.BallotId, rsa.ballotsList[req.BallotId].Rule)
			w.Write([]byte(msg))
			return
		}
		resp.Winner = swf[0]
		resp.Ranking = swf
		serial, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			msg := fmt.Sprintf("error /result: can't serialize response for ballot %s of type %s", req.BallotId, rsa.ballotsList[req.BallotId].Rule)
			w.Write([]byte(msg))
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		w.Write(serial)
		return

	} else {
		var swfVote func(comsoc.Profile) (comsoc.Count, error)
		switch rsa.ballotsList[req.BallotId].Rule {
		case restagent.Borda:
			swfVote = comsoc.BordaSWF
		case restagent.Copeland:
			// Note: Tie-break is applied for Copeland only after SWF calculation, not within the process
			swfVote = comsoc.CopelandSWF
		case restagent.Majority:
			swfVote = comsoc.MajoritySWF
		default:
			w.WriteHeader(http.StatusBadRequest) // 400
			msg := fmt.Sprintf("error /result: type %s is not authorized for ballot %s", rsa.ballotsList[req.BallotId].Rule, req.BallotId)
			w.Write([]byte(msg))
			return
		}

		// Apply tie-break to get the best element and ranking
		var tieBreak = comsoc.TieBreakFactory(rsa.ballotsList[req.BallotId].TieBreak)
		swfFunc := comsoc.SWFFactory(swfVote, tieBreak)
		res, err := swfFunc(rsa.ballotsMap[req.BallotId])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			msg := fmt.Sprintf("error /result: can't process SWF with Tie-break for ballot %s of type %s. "+err.Error(), req.BallotId, rsa.ballotsList[req.BallotId].Rule)
			w.Write([]byte(msg))
			return
		}
		resp.Winner = res[0]
		resp.Ranking = res

		serial, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError) // 500
			msg := fmt.Sprintf("error /result: can't serialize response for ballot %s of type %s", req.BallotId, rsa.ballotsList[req.BallotId].Rule)
			w.Write([]byte(msg))
			return
		}
		w.WriteHeader(http.StatusOK) // 200
		w.Write(serial)
	}
}
