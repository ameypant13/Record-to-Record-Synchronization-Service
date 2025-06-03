package main

import (
	"fmt"
	"github.com/ameypant13/Record-to-Record-Synchronization-Service/pkg/dispatcher"
	schema_validator "github.com/ameypant13/Record-to-Record-Synchronization-Service/pkg/schema-validator"
	sync_worker "github.com/ameypant13/Record-to-Record-Synchronization-Service/pkg/sync-worker"
	"os"
	"time"
)

// +go:embed schema/internal.json
// +go:embed schema/external.json

func main() {
	// Load schemas
	internalContactSchema, err := LoadSchema("schema/internal.json")
	if err != nil {
		fmt.Println("Error loading internal schema:", err)
		return
	}
	externalContactSchema, err := LoadSchema("schema/external.json")
	if err != nil {
		fmt.Println("Error loading external schema:", err)
		return
	}
	worker := &sync_worker.SyncWorker{}
	// Build demo jobs
	jobs := []sync_worker.SyncJob{
		{
			Name: "good",
			Record: map[string]interface{}{
				"id": "abc1", "first_name": "Alice", "last_name": "Doe",
				"email": "alice@example.com", "status": "Active", "priority": "High",
				"ID": "1234",
			},
			SourceSchema:  internalContactSchema,
			DestSchema:    externalContactSchema,
			MappingConfig: schema_validator.InternalToExternalContactConfig,
			Operation:     "create",
		},
		{
			Name: "bad_input_missing_field",
			Record: map[string]interface{}{
				"id": "abc2", "first_name": "Bob",
				// missing last_name!
				"email": "bob@example.com", "status": "Active", "priority": "Medium",
			},
			SourceSchema:  internalContactSchema,
			DestSchema:    externalContactSchema,
			MappingConfig: schema_validator.InternalToExternalContactConfig,
			Operation:     "update",
		},
		{
			Name: "bad_input_enum",
			Record: map[string]interface{}{
				"id": "abc3", "first_name": "Dan", "last_name": "Smith",
				"email": "dan@example.com", "status": "Unknown", "priority": "Low",
			},
			SourceSchema:  internalContactSchema,
			DestSchema:    externalContactSchema,
			MappingConfig: schema_validator.InternalToExternalContactConfig,
			Operation:     "update",
		},
	}

	// 2 API calls per second, allow burst of 2, queue holds 10
	dispatch := dispatcher.NewDispatcher(worker, 2, 2, 10)

	// Prepare jobs...
	for _, job := range jobs {
		if err := dispatch.Submit(job); err != nil {
			fmt.Println("Job queue full or error:", err)
		}
	}

	time.Sleep(3 * time.Second) // Wait for jobs to finish
	dispatch.Shutdown()
}

func LoadSchema(filename string) ([]byte, error) {
	return os.ReadFile(filename) // os.ReadFile is preferred in 1.16+
}
