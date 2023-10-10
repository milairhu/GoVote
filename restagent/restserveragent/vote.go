package restserveragent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent"
)

// Fonctions qui traitent l'appel à l'API REST pour voter:
// http://localhost:8080/vote

// Décode la requête
func (*RestServerAgent) decodeVoteRequest(r *http.Request) (req restagent.RequestVote, err error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	err = json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		fmt.Println("Erreur de décodage de la requête /vote : ", err)

	}
	return
}

func (rsa *RestServerAgent) doVote(w http.ResponseWriter, r *http.Request) {
	rsa.Lock()
	defer rsa.Unlock()
	// vérification de la méthode de la requête
	if !rsa.checkMethod("POST", w, r) {
		return
	}

	// décodage de la requête
	req, err := rsa.decodeVoteRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Println("Serveur recoit : ", r.URL, req)
	//Vérifie que le ballot existe
	_, found := rsa.ballotsList[req.BallotId]
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		msg := fmt.Sprintf("error /vote : ballot %s does not exist", req.BallotId)
		w.Write([]byte(msg))
		return
	}

	//vérifie que l'agent n'a pas déjà voté
	fmt.Println("Ont voté  : ", rsa.ballotsList[req.BallotId].HaveVoted)
	for _, v := range rsa.ballotsList[req.BallotId].HaveVoted {
		if v == req.AgentId {
			w.WriteHeader(http.StatusBadRequest)
			msg := fmt.Sprintf("error /vote : agent %s has already voted for ballot %s", req.AgentId, req.BallotId)
			w.Write([]byte(msg))
			return
		}
	}

	//Vérifie que l'agent a le droit de voter
	var canVote bool
	for _, v := range rsa.ballotsList[req.BallotId].VoterIds {
		if v == req.AgentId {
			canVote = true
			break
		}
	}
	if !canVote {
		w.WriteHeader(http.StatusBadRequest)
		msg := fmt.Sprintf("error /vote : agent %s is not allowed to vote for ballot %s", req.AgentId, req.BallotId)
		w.Write([]byte(msg))
		return
	}

	//Vérifie que la date de fin est n'est pas passée
	if rsa.ballotsList[req.BallotId].Deadline.Before(time.Now()) {
		w.WriteHeader(http.StatusBadRequest)
		msg := fmt.Sprintf("error /result : ballot %s is already fnished", req.BallotId)
		w.Write([]byte(msg))
		return
	}

	//Enregistre le vote pour le ballot
	rsa.ballotsMap[req.BallotId] = append(rsa.ballotsMap[req.BallotId], req.Prefs)

	//Enregistre que l'agent a voté : TODO ne marche pas
	for i := 0; i < len(rsa.ballotsList[req.BallotId].HaveVoted); i++ {
		if rsa.ballotsList[req.BallotId].HaveVoted[i] == "" {
			rsa.ballotsList[req.BallotId].HaveVoted[i] = req.AgentId
			break
		}
	}

	w.WriteHeader(http.StatusOK)
	serial, err := json.Marshal(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err.Error())
		return
	}

	w.Write(serial)
}