package main

import "gitlab.utc.fr/milairhu/ia04-api-rest/restagent/instances"

/**
* This command launches a server and a fleet of voting agents
* to calculate the results of the example from slides 12 and 13, chapter 2
* of the course.
**/

func main() {
	instances.LaunchAgents(2, 6, 3, instances.InitChap3Diap12_13)
}
