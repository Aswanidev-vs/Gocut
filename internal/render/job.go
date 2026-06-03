package render

type JobStatus string

const (
	JobQueued    JobStatus = "queued"
	JobRendering JobStatus = "rendering"
	JobDone      JobStatus = "done"
	JobError     JobStatus = "error"
	JobCancelled JobStatus = "cancelled"
)
