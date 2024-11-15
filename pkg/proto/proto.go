package proto

// Task represents the Proof of Work challenge sent from the server to the client.
type Task struct {
	Prefix     []byte `json:"prefix"`
	Difficulty uint8  `json:"difficulty"`
}

// Solution represents the client's solution to the Proof of Work challenge.
type Solution struct {
	Nonce []byte `json:"nonce"`
}

// Result represents the server's response to the client's solution.
type Result struct {
	// Error contains an error message if the solution was incorrect.
	Error *string `json:"error,omitempty"`
	// Quote contains a wisdom quote if the solution was correct.
	Quote *Quote `json:"quote,omitempty"`
}

type Quote struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}
