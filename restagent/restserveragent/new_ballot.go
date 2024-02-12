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

// Functions that handle the call to the REST API to create a ballot:
// http://localhost:8080/new_ballot

// Decode the request
func (*RestServerAgent) decodeNewBallotRequest(r *http.Request) (req restagent.RequestNewBallot, err error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		fmt.Println("Error decoding request /new_ballot: ", err)
		return
	}
	return
}

// Perform several checks on the provided ballot
func checkBallot(req restagent.RequestNewBallot) (err error) {
	// Check that the date format is correct
	_, err = time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		return fmt.Errorf("deadline")
	}

	// Check that the type of ballot is allowed
	var authorized = false
	for _, v := range restagent.Rules {
		if v == req.Rule {
			authorized = true
			break
		}
	}
	if !authorized {
		return fmt.Errorf("rule")
	}

	// Check that the alternatives are consistent with the tie-break
	if req.Alts < 1 {
		return fmt.Errorf("alts")
	}

	// Check that the tie-break is consistent with the alternatives
	// Note: since the Tie-break is not used for Condorcet, we do not check if it is consistent
	if req.Rule != restagent.Condorcet {
		if req.TieBreak == nil || len(req.TieBreak) != req.Alts {
			return fmt.Errorf("tiebreak")
		} else {
			// Check for duplicates or aberrant values in the tie-break
			list := make([]comsoc.Alternative, len(req.TieBreak))
			copy(list, req.TieBreak)
			sort.Slice(list, func(i, j int) bool { return list[i] < list[j] })
			for i := 0; i < len(req.TieBreak)-1; i++ {
				if list[i]+1 != list[i+1] {
					return fmt.Errorf("tiebreak")
				}
			}
		}
	}

	return
}

func (rsa *RestServerAgent) doCreateNewBallot(w http.ResponseWriter, r *http.Request) {
	rsa.Lock()
	defer rsa.Unlock()
	// Check the request method
	if !rsa.checkMethod("POST", w, r) {
		return
	}

	// Decode the request
	req, err := rsa.decodeNewBallotRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err.Error())
		return
	}

	err = checkBallot(req)

	if err != nil {
		switch err.Error() {
		case "deadline":
			w.WriteHeader(http.StatusBadRequest)
			msg := fmt.Sprintf("error /new_ballot: deadline %s is not in the right format", req.Deadline)
			w.Write([]byte(msg))
			return
		case "rule":
			w.WriteHeader(http.StatusNotImplemented)
			msg := fmt.Sprintf("error /new_ballot: rule %s is not implemented", req.Rule)
			w.Write([]byte(msg))
			return
		case "alts":
			w.WriteHeader(http.StatusBadRequest)
			msg := fmt.Sprintf("error /new_ballot: number of alternatives %d should be >= 1", req.Alts)
			w.Write([]byte(msg))
			return
		case "tiebreak":
			w.WriteHeader(http.StatusBadRequest)
			msg := fmt.Sprintf("error /new_ballot: given tie-break %d is invalid or doesn't match #alts %d", req.TieBreak, req.Alts)
			w.Write([]byte(msg))
			return
		}
	}

	// Register the new ballot
	var ballotId string = fmt.Sprintf("ballot%d", rsa.countBallot)
	rsa.countBallot++
	rsa.ballotsList[ballotId], err = restagent.NewBallot(ballotId, req.Rule, req.Deadline, req.VoterIds, req.Alts, req.TieBreak)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf("error /new_ballot: can't create ballot %s. "+err.Error(), ballotId)
		w.Write([]byte(msg))
		return
	}
	var resp restagent.ResponseNewBallot = restagent.ResponseNewBallot{BallotId: ballotId}

	serial, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprint("error /new_ballot: serialization of response:", err.Error())
		w.Write([]byte(msg))
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(serial)
	return
}
