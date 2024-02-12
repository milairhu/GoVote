package main

import "gitlab.utc.fr/milairhu/ia04-api-rest/restagent/instances"

/**
* This command launches a server and a fleet of voting agents
* to calculate the results of 2 approval polls, one involving the use of Tie-Break, the other not.
**/

func main() {
	instances.LaunchAgents(2, 5+4+3+3, 5, instances.InitApproval)
}
