package restserveragent

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent"
	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent/comsoc"
	"gitlab.utc.fr/milairhu/ia04-api-rest/restagent/endpoints"
)

// RestServerAgent handles HTTP requests for a REST API server.
type RestServerAgent struct {
	sync.Mutex                              // Requests must be processed sequentially, as some requests vote while others request results
	addr        string                      // Server address (ip:port)
	ballotsMap  map[string]comsoc.Profile   // Associates a ballot ID with its profile
	ballotsList map[string]restagent.Ballot // Associates a ballot ID with its Ballot object
	countBallot int                         // Ballot counter (for generating IDs)
}

// NewRestServerAgent creates a new RestServerAgent instance with the given address.
func NewRestServerAgent(addr string) *RestServerAgent {
	b := make(map[string]comsoc.Profile, 0)
	l := make(map[string]restagent.Ballot, 0)
	return &RestServerAgent{addr: addr, ballotsMap: b, ballotsList: l, countBallot: 1}
}

// checkMethod tests the method (GET, POST, ...) of the request.
func (rsa *RestServerAgent) checkMethod(method string, w http.ResponseWriter, r *http.Request) bool {
	if r.Method != method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "method %q not allowed", r.Method)
		return false
	}
	return true
}

// Start starts the REST server.
func (rsa *RestServerAgent) Start() {
	// Create the multiplexer
	mux := http.NewServeMux()
	mux.HandleFunc(endpoints.Results, rsa.doCalcResult)
	mux.HandleFunc(endpoints.Vote, rsa.doVote)
	mux.HandleFunc(endpoints.NewBallot, rsa.doCreateNewBallot)

	// Create the HTTP server
	s := &http.Server{
		Addr:           rsa.addr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20}

	// Start the server
	log.Println("Listening on", rsa.addr)
	go log.Fatal(s.ListenAndServe())
}
