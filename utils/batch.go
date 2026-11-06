package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
)

// AWS Batch commands. Output is one line per item:
//
//	ComputeEnv: <name>, <type>, <state>, <status>, <resource type>, <min>/<desired>/<max> vCPUs
//	JobQueue: <name>, <state>, <status>, priority <n>
//	Job: <id>, <name>, <status>, <status reason>
//	ComputeEnvState: <name>, <ENABLED|DISABLED>
//	Submitted: <job id>, <job name>
//
// The status reason is last because it may contain commas.
// All functions return an error instead of printing it, so the caller can
// exit with a non-zero status.

// DescribeBatch prints the compute environments and job queues, or the jobs
// in jobQueue when it is not empty.
func DescribeBatch(svc *batch.Batch, jobQueue string) error {
	if jobQueue != "" {
		return ListBatchJobs(svc, jobQueue)
	}
	if err := ListBatchComputeEnvironments(svc); err != nil {
		return err
	}
	return ListBatchJobQueues(svc)
}

// ListBatchComputeEnvironments prints all compute environments.
func ListBatchComputeEnvironments(svc *batch.Batch) error {
	return svc.DescribeComputeEnvironmentsPages(&batch.DescribeComputeEnvironmentsInput{},
		func(page *batch.DescribeComputeEnvironmentsOutput, _ bool) bool {
			for _, ce := range page.ComputeEnvironments {
				resType, vcpus := "-", "-"
				if cr := ce.ComputeResources; cr != nil {
					resType = strOr(cr.Type, "-")
					vcpus = fmt.Sprintf("%d/%d/%d vCPUs",
						aws.Int64Value(cr.MinvCpus), aws.Int64Value(cr.DesiredvCpus), aws.Int64Value(cr.MaxvCpus))
				}
				fmt.Printf("ComputeEnv: %s, %s, %s, %s, %s, %s\n",
					strOr(ce.ComputeEnvironmentName, "None"), strOr(ce.Type, "-"),
					strOr(ce.State, "-"), strOr(ce.Status, "-"), resType, vcpus)
			}
			return true
		})
}

// ListBatchJobQueues prints all job queues.
func ListBatchJobQueues(svc *batch.Batch) error {
	return svc.DescribeJobQueuesPages(&batch.DescribeJobQueuesInput{},
		func(page *batch.DescribeJobQueuesOutput, _ bool) bool {
			for _, q := range page.JobQueues {
				fmt.Printf("JobQueue: %s, %s, %s, priority %d\n",
					strOr(q.JobQueueName, "None"), strOr(q.State, "-"),
					strOr(q.Status, "-"), aws.Int64Value(q.Priority))
			}
			return true
		})
}

// ListBatchJobs prints the jobs in a queue for every status. ListJobs only
// returns RUNNING jobs when no status is given, so each status is queried.
// Batch keeps finished (SUCCEEDED/FAILED) jobs for about 7 days.
func ListBatchJobs(svc *batch.Batch, jobQueue string) error {
	// ListJobs returns nothing for an unknown queue, so check it exists first.
	res, err := svc.DescribeJobQueues(&batch.DescribeJobQueuesInput{JobQueues: []*string{aws.String(jobQueue)}})
	if err != nil {
		return err
	}
	if len(res.JobQueues) == 0 {
		return fmt.Errorf("job queue %s not found", jobQueue)
	}
	for _, status := range batch.JobStatus_Values() {
		input := &batch.ListJobsInput{
			JobQueue:  aws.String(jobQueue),
			JobStatus: aws.String(status),
		}
		err := svc.ListJobsPages(input, func(page *batch.ListJobsOutput, _ bool) bool {
			for _, j := range page.JobSummaryList {
				fmt.Printf("Job: %s, %s, %s, %s\n",
					strOr(j.JobId, "None"), strOr(j.JobName, ""),
					strOr(j.Status, status), strOr(j.StatusReason, ""))
			}
			return true
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// SetBatchComputeEnvironmentState enables or disables a compute environment.
// A Fargate environment has no instances to start: enabling it only lets
// jobs in its queues run.
func SetBatchComputeEnvironmentState(svc *batch.Batch, name string, enable bool) error {
	if name == "" {
		return fmt.Errorf("compute environment name (-n) is required")
	}
	state := batch.CEStateDisabled
	if enable {
		state = batch.CEStateEnabled
	}
	_, err := svc.UpdateComputeEnvironment(&batch.UpdateComputeEnvironmentInput{
		ComputeEnvironment: aws.String(name),
		State:              aws.String(state),
	})
	if err != nil {
		return err
	}
	fmt.Printf("ComputeEnvState: %s, %s\n", name, state)
	return nil
}

// SubmitBatchJob submits a job to jobQueue using jobDefinition (name, or
// name:revision; the latest active revision when no revision is given).
// jobName defaults to "<job definition name>-<timestamp>".
func SubmitBatchJob(svc *batch.Batch, jobQueue, jobDefinition, jobName string) error {
	if jobQueue == "" || jobDefinition == "" {
		return fmt.Errorf("job queue (-n) and job definition (-jobdef) are required")
	}
	if jobName == "" {
		defName, _, _ := strings.Cut(jobDefinition, ":")
		jobName = defName + "-" + time.Now().Format("20060102-150405")
	}
	res, err := svc.SubmitJob(&batch.SubmitJobInput{
		JobQueue:      aws.String(jobQueue),
		JobDefinition: aws.String(jobDefinition),
		JobName:       aws.String(jobName),
	})
	if err != nil {
		return err
	}
	fmt.Printf("Submitted: %s, %s\n", strOr(res.JobId, "None"), strOr(res.JobName, jobName))
	return nil
}
