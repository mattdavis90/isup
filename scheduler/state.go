package scheduler

type State int

const (
	OkState State = iota
	PendingState
	AlertingState
	NoDataState
)

func (s State) String() string {
	return [...]string{"Ok", "Pending", "Alerting", "No_Data"}[s]
}
