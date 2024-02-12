package restagent

import (
	"time"

	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent/comsoc"
)

// Types used for the /new_ballot request
type Ballot struct {
	BallotId   string               // Ballot identifier
	Rule       string               // Voting method
	Deadline   time.Time            // Voting deadline
	VoterIds   []string             // List of agents eligible to vote
	Alts       int                  // Number of alternatives (from 1 to Alts)
	TieBreak   []comsoc.Alternative // Preference order of alternatives in case of a tie
	HaveVoted  []string             // Names of agents who have voted
	Thresholds map[string]int       // Contains the thresholds of each voter (for approval voting)
}

// Constructor for a Ballot
func NewBallot(ballotId string, rule string, deadline string, voterIds []string, alts int, tieBreak []comsoc.Alternative) (Ballot, error) {
	// Check the date format
	date, err := time.Parse(time.RFC3339, deadline)
	if err != nil {
		return Ballot{}, err
	}
	haveVoted := make([]string, len(voterIds))
	thresholds := make(map[string]int)
	return Ballot{
		BallotId:   ballotId,
		Rule:       rule,
		Deadline:   date,
		VoterIds:   voterIds,
		Alts:       alts,
		TieBreak:   tieBreak,
		HaveVoted:  haveVoted,
		Thresholds: thresholds,
	}, nil
}

type RequestNewBallot struct {
	Rule     string               `json:"rule"`      // Voting method
	Deadline string               `json:"deadline"`  // Voting deadline
	VoterIds []string             `json:"voter-ids"` // List of agents eligible to vote
	Alts     int                  `json:"#alts"`     // Number of alternatives (from 1 to Alts)
	TieBreak []comsoc.Alternative `json:"tie-break"` // Preference order of alternatives in case of a tie
}

type ResponseNewBallot struct {
	// Object returned if code 201
	BallotId string `json:"ballot-id"` // Id of the created ballot
}

// Type used for the /vote request
type RequestVote struct {
	AgentId  string               `json:"agent-id"`  // Id of the voting agent
	BallotId string               `json:"ballot-id"` // Id of the ballot being voted on
	Prefs    []comsoc.Alternative `json:"prefs"`     // Ordered preferences of the voting agent
	Options  []int                `json:"options"`   // Used for the threshold in approval voting
}

// Types used for the /result request

type RequestResult struct {
	BallotId string `json:"ballot-id"` // Id of the ballot for which the result is requested
}

type ResponseResult struct {
	// Object returned if code 200
	Winner  comsoc.Alternative   `json:"winner"`            // Winning alternative
	Ranking []comsoc.Alternative `json:"ranking,omitempty"` // Ranking of alternatives (Optional field)
}
