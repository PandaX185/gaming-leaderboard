package consts

import "time"

const (
	ScoreUpdatedEvent  = "ScoreUpdated"
	ScoreEvents        = "score_events"
	ScoreConsumerGroup = "score-worker-group"
	ScoreReclaimTime   = time.Minute * 1
)
