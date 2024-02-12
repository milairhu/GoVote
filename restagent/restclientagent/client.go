package restclientagent

import (
	"fmt"
	"log"
	"time"

	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent"
)

/******************  Client Agent Body ******************/
type RestClientAgentBase struct {
	Id   string          // Agent ID
	url  string          // Server URL
	cin  <-chan []string // Channel to receive list of ballots
	cout chan<- string   // Channel to communicate the name of its ballot
}

/******************  Agent Creating a Ballot and Calculating Result ******************/
type RestClientBallotAgent struct {
	RestClientAgentBase                            // Basic client agent attributes
	ReqNewBallot        restagent.RequestNewBallot // Request for creating a new ballot
}

// Constructor for an agent creating a ballot
func NewRestClientBallotAgent(id string, url string, reqNewBallot restagent.RequestNewBallot, cin <-chan []string, cout chan<- string) *RestClientBallotAgent {
	return &RestClientBallotAgent{
		RestClientAgentBase{id, url, cin, cout},
		reqNewBallot,
	}
}

// Main method of the ballot creating agent
func (rcba *RestClientBallotAgent) Start() {
	// Step 1: Creating the ballot
	createdBallot, err := rcba.doRequestNewBallot(rcba.ReqNewBallot)
	if err != nil {
		log.Printf(rcba.Id, " error: ", err.Error())
	} else {
		// log.Printf("/new_Ballot by [%s] created successfully: %s\n", rcba.id, createdBallot.BallotId)
	}
	// Step 2: Sending its ballot to the main goroutine
	rcba.cout <- createdBallot.BallotId

	// Step 3: Waiting for all agents to finish voting, signaled by the main goroutine
	<-rcba.cin

	time.Sleep(6 * time.Second)

	// Step 4: Retrieving the result of each ballot
	res, err := rcba.doRequestResults(createdBallot.BallotId)
	if err != nil {
		log.Printf(rcba.Id, "error: ", err.Error())
	} else {
		Affichage(createdBallot.BallotId, rcba.ReqNewBallot.Rule, len(rcba.ReqNewBallot.VoterIds), res)
	}
}

/****************** Voting Agent ******************/
type RestClientVoteAgent struct {
	RestClientAgentBase                       // Basic client agent attributes
	ReqVote             restagent.RequestVote // Request for voting
}

// Constructor for a voting agent
func NewRestClientVoteAgent(id string, url string, reqVote restagent.RequestVote, cin <-chan []string, cout chan<- string) *RestClientVoteAgent {
	return &RestClientVoteAgent{
		RestClientAgentBase{id, url, cin, cout},
		reqVote,
	}
}

// Main method of the voting agent
func (rcva *RestClientVoteAgent) Start() {
	// Step 1: Waiting to receive the list of ballots sent by the main goroutine
	listBallots := <-rcva.cin

	// Step 2: Voting in each ballot
	for _, ballot := range listBallots {
		rcva.ReqVote.BallotId = ballot // Set the ballot ID in the request
		err := rcva.doRequestVote(rcva.ReqVote)
		if err != nil {
			log.Printf(rcva.Id, " error: ", err.Error())
		}
	}
	// Step 3: Sending a message to the main goroutine to indicate completion
	rcva.cout <- "fin"
}

// Display results
func Affichage(id string, rule string, nbVoters int, res restagent.ResponseResult) {
	if rule != "condorcet" {
		fmt.Printf("=============================== RESULTS FOR BALLOT %s ===============================\nBALLOT TYPE: %s\nNUMBER OF VOTERS: %d\nWINNER: %d\nRANKING: %v\n",
			id, rule, nbVoters, res.Winner, res.Ranking)
	} else {
		fmt.Printf("=============================== RESULTS FOR BALLOT %s ===============================\nBALLOT TYPE: %s\nNUMBER OF VOTERS: %d\nWINNER: %d\n",
			id, rule, nbVoters, res.Winner)
	}
}
