package events

type InconsistentEvent struct {
	ID int64
	// 用来修什么，取值为 SRC，意味着，以源表为准，取值为 DST，以目标表为准
	Direction string
	Type      string
}

const (
	InconsistentEventTypeTargetMissing = "target_missing"
	InconsistentEventTypeNEQ           = "new"
	InconsistentEventTypeBaseMissing   = "base_missing"
)
