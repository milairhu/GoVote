# Web Services REST for Voting Methods and Ballots

By Hugo MILAIR, Damien VAURS
2023-10-30

## Introduction

This project was led as part of the IA04 course at UTC, taught by Pr. Sylvain Lagrue. The goal was to create a REST API for voting methods and ballots and to test it with a set of agents. Some agents create ballots, manage them and process the result while other agents, the voters, submit their votes. Time for ballot and deadline management is also implemented.

## Installation et lancement

The project can be cloned via the following command:

    https://gitlab.utc.fr/milairhu/ia04-api-rest.git

Without cloning the project, the different executable functions, located in the folder /restagent/cmd, can be downloaded via the following command:

    go get gitlab.utc.fr/milairhu/ia04-api-rest/restagent/cmd/launch-10-generated-agents@latest
where *launch-10-generated-agents* can be replaced by the name of any other executable cmd folder. The use of @latest is only mandatory if the active folder is not in a Go module.

To run a program via *go run*, simply go to the *cmd* folder and execute the following command:

    go run launch-10-generated-agents.go

To run an executable installed via *go install*, simply navigate to the *go/bin* folder, typically in the home directory, and run the command like this:

    ./launch-10-generated-agents

Note that the program may fail, as the server agent occasionally fails to connect to port 8080. In this case, you need to restart the program.

## General Structure (Packages)

### cmd Folder

The *restagent/cmd/* folder contains all the executable files used to test the entire project. Some of these files revisit the examples seen in class.

launch-10-generated

The *restagent/cmd/* folder contains all the executable files used to test the entire project. Some of these files revisit the examples seen in class.

- *launch-10-generated-agents.go*: launches 10 randomly generated voting agents to test each implemented voting method and some edge cases.
- *launch-x-generated-agents.go*: Similar to the previous one, except that the user is asked to provide the number of voters, ballots, and alternatives. Handy for testing scenarios with a very large number of agents. Note: no edge cases are generated (expired deadline, voter not entitled to vote, etc.). The voting methods for each ballot are chosen randomly.
- *launch-approval.go*, *launch-condorcet.go*, and *launch-stv.go*: allow testing the Approval, Condorcet, and STV methods with and without the need for tie-break, as their manipulation differs from other methods.
- *launch-rsagt.go*: launches a REST server that handles incoming requests on port 8080. This is the command to run if the user wants to test the API via a tool like Postman.
- *launch-rcagt.go*: launches a REST client that sends requests to the previously launched REST server. It starts a simple ballot creator agent and a voting agent.
- The commands in the files *launch-chap2-diapX.go* allow testing the examples seen in class.

### Package comsoc

This package (*directory /restagent/comsoc/*) contains all the classes, methods, and types related to ballot management. It includes functions for calculating SWF and SCF for methods like Borda, Condorcet, Copeland, etc.

It also includes a number of utility functions grouped in the *file /comsoc/basics.go*.

Finally, the *file /comsoc/tiebreak.go* contains **factory** functions for creating tie-break functions for different methods. Only tie-breaks for STV and Approval had to be implemented manually, as their use differs from other methods.

### Package endpoints

Endpoints (*directory /restagent/endpoints/*) is a package consisting of a single *file /endpoints/endpoints.go* whose purpose is to define certain constants used throughout the project. It contains elements for constructing URLs for HTTP requests.

### Package instances

The instances package (*directory restagent/instances/*) contains all the initialization files for the executables in the cmd folder and have the same names as those executables.

The *init-....go* files each contain a function that instantiates voting agents and ballot managers for the associated executable.

The *launch-agents.go* file contains the abstraction of the function launching the agents. Since originally, for each executable, the **main** functions were almost identical, it was decided to abstract them in this file.

### Package restclientagent

In this package (*directory /restagent/restclientagent/*), you find the definition of client-side agents (voters and ballot managers) as well as the methods used to make various HTTP requests.

### Package restserveragent

This package (*directory /restagent/restserveragent/*), similar in design to the previous one, defines all the classes and methods on the server side. Its functions allow the server agent to communicate with client agents from the *restclientagent* package via HTTP requests.

## Package restagent

The restagent package, located at the root of the project, defines a number of types (*file /types.go*) and constants (*file /rule.go*) used by client and server agents.

## Remarks

- For the Approval voting method, a new tie-break method had to be created to account for thresholds (function *MakeApprovalRankingWithTieBreak()* in the file */comsoc/tiebreak.go*).
- The same applies to the STV method, which requires a separate tie-break function because the tie-breaking occurs within the SWF calculation function itself, rather than afterwards (function *STV_SWF_TieBreak* in the file */comsoc/tiebreak.go*).
- When creating a ballot, we do not use log.Fatal because we want the agent to continue its tasks even if an error is encountered.
- Checking the consistency of thresholds provided at the time of result calculation (file */restserveragent/result.go*) offers security advantages but penalizes performance, as these thresholds are already checked upon receiving the vote.
- There is a question about whether it is beneficial (or not) to check the presence of a threshold in the context of an Approval voting method. Is the absence of a threshold an error? Or does it mean that all alternatives are counted or none? It was decided to consider the absence of a threshold as an error.
- In cases where no votes are submitted (file */restserveragent/result.go*), we decide to return a result rather than an error. This result is determined by the tie-break provided when creating the ballot.
- The Condorcet method does not require the use of a tie-break management function. Either there is a single winner or there isn't.
- The number of alternatives in the *file /cmd/launch-rcagt.go* was arbitrarily set to 5 but is adjustable.
- The "voters profile" displayed has the same format as those seen in class. The first line indicates the number of voters with this preference order, indicated by the column below.

Hugo MILAIR,
Damien VAURS,
Semester A23, 2023-10-30
