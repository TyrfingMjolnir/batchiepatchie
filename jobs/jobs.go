/*
Package jobs implements the basic Job structure and related functionality
*/
package jobs

import (
	"time"
)

// Job Status Constants
const (
	StatusFailed    = "FAILED"
	StatusPending   = "PENDING"
	StatusRunnable  = "RUNNABLE"
	StatusRunning   = "RUNNING"
	StatusStarting  = "STARTING"
	StatusSubmitted = "SUBMITTED"
	StatusSucceeded = "SUCCEEDED"
)

// StatusList is a list of all possible job statuses
var StatusList = [...]string{
	StatusFailed,
	StatusPending,
	StatusRunnable,
	StatusRunning,
	StatusStarting,
	StatusSubmitted,
	StatusSucceeded,
}

type JobStatus struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

type Job struct {
	Id                   string     `json:"id"`
	Name                 string     `json:"name"`
	Status               string     `json:"status"`
	Description          string     `json:"desc"`
	LastUpdated          time.Time  `json:"last_updated"`
	JobQueue             string     `json:"job_queue"`
	Image                string     `json:"image"`
	CreatedAt            time.Time  `json:"created_at"`
	StoppedAt            *time.Time `json:"stopped_at"`
	VCpus                int64      `json:"vcpus"`
	Memory               int64      `json:"memory"`
	Timeout              int        `json:"timeout"`
	CommandLine          string     `json:"command_line"`
	StatusReason         *string    `json:"status_reason"`
	RunStartTime         *time.Time `json:"run_start_time"`
	ExitCode             *int64     `json:"exitcode"`
	LogStreamName        *string    `json:"log_stream_name"`
	TerminationRequested bool       `json:"termination_requested"`
	TaskARN              *string    `json:"task_arn"`
	InstanceID           *string    `json:"instance_id"`
	PublicIP             *string    `json:"public_ip"`
	PrivateIP            *string    `json:"private_ip"`
}

// Options is the query options for the Find method to use
type Options struct {
	Search  string
	Limit   int
	Offset  int
	Queues  []string
	SortBy  string
	SortAsc bool
	Status  []string
}

// KillTaskID is a struct to handle JSON request to kill a task
type KillTaskID struct {
	ID string `json:"id" form:"id" query:"id"`
}

// FinderStorer is an interface that can both save and retrieve jobs
type FinderStorer interface {
	Finder
	Storer

	// Methods to get information about Job Queues
	ListActiveJobQueues() ([]string, error)
	ListForcedScalingJobQueues() ([]string, error)

	ActivateJobQueue(string) error
	DeactivateJobQueue(string) error
}

// Finder is an interface to find jobs in a database/store
type Finder interface {
	// Find finds a jobs matching the query
	Find(opts *Options) ([]*Job, error)

	// FindOne finds a job matching the query
	FindOne(query string) (*Job, error)

	// FindTimedoutJobs finds all job IDs that should have timed out by now
	FindTimedoutJobs() ([]string, error)

	// Simple endpoint that returns a string for job status.
	GetStatus(jobid string) (*JobStatus, error)
}

// Storer is an interface to save jobs in a database/store
type Storer interface {
	// Store saves a job
	Store(job []*Job) error

	// Gives the store a chance to stale jobs we no longer know about
	// The argument is a set (value is ignored) of all known job_ids currently by AWS Batch
	StaleOldJobs(map[string]bool) error

	// Finds estimated load per job queue
	EstimateRunningLoadByJobQueue([]string) (map[string]RunningLoad, error)

	// Update compute environment logs
	UpdateComputeEnvironmentsLog([]ComputeEnvironment) error

	// Update job summaries
	UpdateJobSummaryLog([]JobSummary) error

	// Mark on job that we requested it to be terminated
	UpdateJobLogTerminationRequested(string) error

	// Updates information on task arns and ec2 metadata
	UpdateTaskArnsInstanceIDs(map[string]Ec2Info, map[string]string) error

	// Updates information on EC2 instances running on ECS
	UpdateECSInstances(map[string]Ec2Info, map[string][]string) error

	// Gets alive EC2 instances (according to database)
	GetAliveEC2Instances() ([]string, error)

	// Gets all instance IDs that have jobs stuck in "STARTING" status
	GetStartingStateStuckEC2Instances() ([]string, error)

	// Subscribes to updates about a job status. (see more info on this
	// function in postgres_store.go)
	SubscribeToJobStatus(jobID string) (<-chan Job, func())
}

// Killer is an interface to kill jobs in the queue
type Killer interface {
	// KillOne kills a job matching the query
	KillOne(jobID string, reason string, store Storer) error

	// Kills jobs and instances that are stuck in STARTING status
	KillInstances(instances []string) error
}

// This structure describes how many vcpus and memory the currently queued jobs require
type RunningLoad struct {
	WantedVCpus  int64
	WantedMemory int64
}

type ComputeEnvironment struct {
	Name        string
	WantedvCpus int64
	MinvCpus    int64
	MaxvCpus    int64
	State       string
	ServiceRole string
}

type JobSummary struct {
	JobQueue  string
	Submitted int64
	Pending   int64
	Runnable  int64
	Starting  int64
	Running   int64
}

type Ec2Info struct {
	PrivateIP             *string
	PublicIP              *string
	AMI                   string
	ComputeEnvironmentARN string
	ECSClusterARN         string
	AvailabilityZone      string
	SpotInstanceRequestID *string
	InstanceType          string
	LaunchedAt            *time.Time
}
