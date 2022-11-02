package api

// HTTP Methods
const (
	GET    = "GET"
	POST   = "POST"
	DELETE = "DELETE"
	PUT    = "PUT"
)

// A request object
type HciRequest struct {
	Method   string
	Endpoint string
	Body     []byte
	Options  map[string]string
}
