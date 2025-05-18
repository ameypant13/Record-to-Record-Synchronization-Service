package main

import (
	"fmt"
	schema_validator "github.com/ameypant13/Record-to-Record-Synchronization-Service/pkg/schema-validator"
	"os"
)

// +go:embed schema/internal.json
// +go:embed schema/external.json

func main() {
	internalSchema, err := LoadSchema("schema/internal.json")
	if err != nil {
		panic(err)
	}
	externalSchema, err := LoadSchema("schema/external.json")
	if err != nil {
		panic(err)
	}

	// Now use internalSchema and externalSchema as []byte
	// For example:
	internalInput := map[string]interface{}{
		"id":         "abc123",
		"first_name": "Alice",
		"last_name":  "Doe",
		"email":      "alice@example.com",
		"status":     "Active",
		"priority":   "High",
	}

	// Use the previously defined TransformAndValidate function
	output, err := schema_validator.TransformAndValidate(
		internalInput,
		internalSchema,
		externalSchema,
		schema_validator.InternalToExternalContactConfig,
	)
	if err != nil {
		fmt.Println("Transform failed:", err)
		return
	}
	fmt.Printf("Transformed output: %+v\n", output)

}

func LoadSchema(filename string) ([]byte, error) {
	return os.ReadFile(filename) // os.ReadFile is preferred in 1.16+
}
