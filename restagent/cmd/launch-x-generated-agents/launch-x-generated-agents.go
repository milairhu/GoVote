package main

import (
	"fmt"

	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent/instances"
)

func main() {

	var nbAgents int
	var nbBallot int
	var nbAlts int

	fmt.Println("How many voting agents ?")
	fmt.Scanln(&nbAgents)
	fmt.Println("How many ballots ?")
	fmt.Scanln(&nbBallot)
	fmt.Println("How many alternatives ?")
	fmt.Scanln(&nbAlts)

	instances.LaunchAgents(nbBallot, nbAgents, nbAlts, instances.InitVotingAgents)
}
