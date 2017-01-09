package state

// StateCondition holds information about a specific state condition, and also functions to use that state condition
type StateConditions struct {
	Exit         *ExitCondition         `json:"exit"`
	Timeout      *TimeoutCondition      `json:"timeout"`
	FileMonitors []FileMonitorCondition `json:"filemonitor"`
	Outputs      []OutputCondition      `json:"output"`
}

// Count returns the total number of state conditions we have
func (state *StateConditions) Count() int {
	count := 0
	if state.Exit != nil {
		count++
	}
	if state.Timeout != nil {
		count++
	}
	count += len(state.FileMonitors)
	count += len(state.Outputs)
	return count
}
