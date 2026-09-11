package handlers

// PositionRequest is the shared payload for position update endpoints.
type PositionRequest struct {
	Longitude float64 `json:"long"`
	Latitude  float64 `json:"lat"`
}
