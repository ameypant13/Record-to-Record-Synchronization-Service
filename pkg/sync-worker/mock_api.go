package sync_worker

import "fmt"

func MockExternalAPI(record map[string]interface{}, op string) error {
	fmt.Printf("Mock API [%s]: sending %+v\n", op, record)
	// For demo, randomly fail if, e.g., record has a field "failme": true
	if b, ok := record["failme"].(bool); ok && b {
		return fmt.Errorf("external system rejected record")
	}
	return nil
}
