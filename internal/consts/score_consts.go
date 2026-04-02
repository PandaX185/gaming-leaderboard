package consts

import "time"

const (
	ScoreEvents        = "score_events"
	ScoreConsumerGroup = "score-worker-group"
	ScoreReclaimTime   = time.Minute * 1
)
