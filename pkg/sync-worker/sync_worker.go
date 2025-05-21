package sync_worker

import schema_validator "github.com/ameypant13/Record-to-Record-Synchronization-Service/pkg/schema-validator"

// SyncJob models a sync attempt
type SyncJob struct {
	Name          string                 // Job ID for demo
	Record        map[string]interface{} // The data to sync
	SourceSchema  []byte
	DestSchema    []byte
	MappingConfig schema_validator.TransformConfig
	Operation     string // "create","update","delete" etc
}

// SyncResult models job outcome
type SyncResult struct {
	JobName     string
	Status      string // "success", "transform_fail", etc
	Detail      string
	Transformed map[string]interface{}
}

type SyncWorker struct{}

func (w *SyncWorker) ProcessJob(job SyncJob) SyncResult {
	// 1. Transform & validate
	out, err := schema_validator.TransformAndValidate(
		job.Record,
		job.SourceSchema,
		job.DestSchema,
		job.MappingConfig,
	)
	if err != nil {
		return SyncResult{
			JobName: job.Name,
			Status:  "transform_fail",
			Detail:  err.Error(),
		}
	}

	return SyncResult{
		JobName:     job.Name,
		Status:      "success",
		Transformed: out,
		Detail:      "OK",
	}
}
