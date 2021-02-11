package standard

type ProgressBar interface {
	SetMinProgress(min int)
	MinProgress() int
	SetMaxProgress(max int)
	MaxProgress() int
	SetCurrentProgress(value int)
	CurrentProgress() int
}
