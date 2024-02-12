package main

import (
	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent"
	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent/instances"
)

/**
* This command launches a server and a fleet of 10 voting agents for 8 polls (1 per voting method and 2 witnesses).
* Their preferences (and their threshold) are randomly generated.
* Some polls are set up to trigger errors (deadline passed, deadline too far away, etc.) to test the robustness of the system.
**/

func main() {
	instances.LaunchAgents(len(restagent.Rules)+3, 10, 5, instances.Init10VotingAgents)
}
