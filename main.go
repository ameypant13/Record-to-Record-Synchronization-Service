package main

import (
	"encoding/json"
	"fmt"
	schema_validator "github.com/ameypant13/Record-to-Record-Synchronization-Service/pkg/schema-validator"
	sync_worker "github.com/ameypant13/Record-to-Record-Synchronization-Service/pkg/sync-worker"
	"os"
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

	// Process all jobs
	var results []sync_worker.SyncResult
	for _, job := range jobs {
		res := worker.ProcessJob(job)
		results = append(results, res)
	}
	// Print result summary
	fmt.Println("\n--- RESULTS ---")
	for _, r := range results {
		fmt.Printf("%s: %s (%s)\n", r.JobName, r.Status, r.Detail)
		if r.Status == "success" {
			enc, _ := json.MarshalIndent(r.Transformed, "  ", "  ")
			fmt.Println("  Output:", string(enc))
		}
	}
}

func LoadSchema(filename string) ([]byte, error) {
	return os.ReadFile(filename) // os.ReadFile is preferred in 1.16+
}
