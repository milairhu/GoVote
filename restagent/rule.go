package restagent

//Set of rules considered for the voting system

const Approval = "approval"
const Borda = "borda"
const Condorcet = "condorcet"
const Copeland = "copeland"
const Majority = "majority"
const STV = "stv"

var Rules = []string{Approval, Borda, Condorcet, Copeland, Majority, STV}
