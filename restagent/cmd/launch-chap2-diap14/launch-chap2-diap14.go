package main

import "gitlab.utc.fr/milairhu/ia04-api-rest/restagent/instances"

/**
* This command launches a server and a fleet of voting agents
* to calculate the results of the example slide 14, chapter 2
* of the course.
**/

func main() {
	instances.LaunchAgents(1, 10+6+5, 3, instances.InitChap3Diap14)
}
