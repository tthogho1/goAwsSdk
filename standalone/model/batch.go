package model

// BatchComputeEnv holds one AWS Batch compute environment
type BatchComputeEnv struct {
	Name         string
	Type         string // MANAGED | UNMANAGED
	State        string // ENABLED | DISABLED
	Status       string // VALID | INVALID | CREATING ...
	ResourceType string // EC2 | SPOT | FARGATE | FARGATE_SPOT, "-" if unmanaged
	VCPUs        string // "<min>/<desired>/<max> vCPUs", "-" if unmanaged
}

// BatchJobQueue holds one AWS Batch job queue
type BatchJobQueue struct {
	Name     string
	State    string
	Status   string
	Priority string
}

// BatchJob holds a summary of one AWS Batch job
type BatchJob struct {
	ID     string
	Name   string
	Status string // SUBMITTED | PENDING | RUNNABLE | STARTING | RUNNING | SUCCEEDED | FAILED
	Reason string
}
