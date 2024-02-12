package main

import (
	"math/rand"
	"sync"
	"time"

	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent"
	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent/comsoc"
	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent/endpoints"
	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent/restclientagent"
)

/**
* This command launches a poll creator agent and a voting agent that perform the following commands:
* - POST /new_ballot: creates a poll
* - POST /vote: votes for the poll
* - POST /result: retrieves the result of the poll
* - Displays the result.
* This is not very useful in itself except to test the operation of the REST API and the synchronization between agents.
**/

const nbAlts = 5 //number of alternatives in the preferences
//Note, we chose this number arbitrarily

func main() {

	//Creating the request to create a new poll
	reqNewBallot := restagent.RequestNewBallot{
		Rule:     restagent.Majority,
		Deadline: time.Now().Add(5 * time.Second).Format(time.RFC3339),
		VoterIds: []string{"ag_vote"}, //only one voter because we only launch one client
		Alts:     nbAlts,
		TieBreak: []comsoc.Alternative{1, 2, 3, 4, 5},
	}

	//Creating the request to vote
	intPref := rand.Perm(nbAlts)
	altPref := make([]comsoc.Alternative, nbAlts)
	for i := 0; i < nbAlts; i++ {
		//conversion to []Alternative
		altPref[i] = comsoc.Alternative(intPref[i] + 1)
	}
	reqVote := restagent.RequestVote{
		AgentId:  "ag_vote",
		BallotId: "",
		Prefs:    altPref,
		Options:  nil,
	}

	//Creating the cin and cout channels
	listChannelsIn := make([]chan []string, 2) //Communication to the agents
	channelOut := make(chan string)            //Communication from the agents to the main goroutine

	listChannelsIn[0] = make(chan []string)
	listChannelsIn[1] = make(chan []string)
	//Launching the agents
	agScrutin := restclientagent.NewRestClientBallotAgent("ag_scrut", endpoints.ServerHost+endpoints.ServerPort, reqNewBallot, listChannelsIn[0], channelOut)
	agVote := restclientagent.NewRestClientVoteAgent("ag_vote", endpoints.ServerHost+endpoints.ServerPort, reqVote, listChannelsIn[1], channelOut)

	wg := sync.WaitGroup{}
	wg.Add(2)

	//Launching the agents

	go func() {
		//Launching the poll creator
		defer wg.Done()
		agScrutin.Start()
	}()

	go func() {
		//Launching the voter
		defer wg.Done()
		agVote.Start()
	}()
	//Wait for the receipt of a poll
	ballotId := <-channelOut
	//Send the list of polls (only 1) to the voter
	listChannelsIn[1] <- []string{ballotId}
	//Wait for the receipt of a message announcing the end of the votes
	<-channelOut
	//Send a message to the poll agent to start the calculation of the results
	listChannelsIn[0] <- []string{""}
	wg.Wait()
}
