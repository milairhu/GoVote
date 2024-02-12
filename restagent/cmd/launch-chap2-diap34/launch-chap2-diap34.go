package main

import "gitlab.utc.fr/milairhu/ia04-api-rest/restagent/instances"

/**
* This command launches a server and a fleet of voting agents
* to calculate the results of the example slide 34, chapter 2
* of the course.
**/

func main() {
	instances.LaunchAgents(3, 5+4+2+6+8+2, 4, instances.InitChap3Diap34)
}
