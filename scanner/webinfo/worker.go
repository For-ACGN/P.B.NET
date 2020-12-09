package webinfo

// WorkerStatus contains worker status.
type WorkerStatus struct {
	Idle   int64 // unix timestamp
	Active int64 // unix timestamp
}
