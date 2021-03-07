package common

type State string

const (
	StateClosed State = "closed"
)

type Label string

const (
	LabelUB      Label = "ub"
	LabelInvalid Label = "invalid"
	LabelMerged  Label = "merged"
	LabelError   Label = "error"
	LabelCheck   Label = "check"
	LabelFetch   Label = "fetch"
)
