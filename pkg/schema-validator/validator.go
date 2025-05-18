package schema_validator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/qri-io/jsonschema"
)

type FieldMapping struct {
	SourceField string
	TargetField string

	// For string enum → int code mapping
	ValueStringToInt map[string]int // string input → int output
	ToInt            bool           // map using ValueStringToInt and set as int

	// For int code → string enum mapping (reverse)
	ValueIntToString map[int]string // int input → string output

	// The existing ValueMap, ToBool, FromBoolMap still valid for other cases
	ValueMap    map[string]string // (legacy; string->string value mapping)
	ToBool      bool              // for string->bool
	FromBoolMap map[bool]string   // for bool->string
}

type TransformConfig struct {
	FieldMappings []FieldMapping
}

var InternalToExternalContactConfig = TransformConfig{
	FieldMappings: []FieldMapping{
		{SourceField: "id", TargetField: "contactId"},
		{SourceField: "first_name", TargetField: "givenName"},
		{SourceField: "last_name", TargetField: "familyName"},
		{SourceField: "email", TargetField: "emailAddress"},
		{
			SourceField: "status",
			TargetField: "isActive",
			ValueMap: map[string]string{
				"Active":   "true",
				"Inactive": "false",
			},
			ToBool: true,
		},
		{
			SourceField: "priority",
			TargetField: "priorityCode",
			ValueStringToInt: map[string]int{
				"Low": 1, "Medium": 2, "High": 3,
			},
			ToInt: true,
		},
	},
}

var ExternalToInternalContactConfig = TransformConfig{
	FieldMappings: []FieldMapping{
		{SourceField: "contactId",  TargetField: "id"},
		{SourceField: "givenName",  TargetField: "first_name"},
		{SourceField: "familyName", TargetField: "last_name"},
		{SourceField: "emailAddress", TargetField: "email"},
		{
			SourceField: "isActive",
			TargetField: "status",
			FromBoolMap: map[bool]string{
				true:  "Active",
				false: "Inactive",
			},
		},
		{
		    SourceField: "priorityCode",
		    TargetField: "priority",
		    ValueIntToString: map[int]string{
		        1: "Low", 2: "Medium", 3: "High",
		    },
		},
	},
}

func validateAgainstSchema(data map[string]interface{}, schemaBytes []byte) error {
	rs := &jsonschema.Schema{}
	if err := json.Unmarshal(schemaBytes, rs); err != nil {
		return errors.New("failed to parse schema: " + err.Error())
	}
	docBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var doc interface{}
	if err := json.Unmarshal(docBytes, &doc); err != nil {
		return err
	}
	key_errs, err := rs.ValidateBytes(context.Background(), docBytes)
	if err != nil {
		return fmt.Errorf("jsonschema validate failed: %v", err)
	}
	if len(key_errs) > 0 {
		var sb strings.Builder
		for _, e := range key_errs {
			sb.WriteString("- " + e.Message + "\n")
		}
		return errors.New("schema validation failed:\n" + sb.String())
	}
	return nil
}

// Transform and validate
func TransformAndValidate(
	input map[string]interface{},
	sourceSchema []byte,
	destSchema []byte,
	config TransformConfig,
) (map[string]interface{}, error) {
	// 1. Validate input against the source schema
	if err := validateAgainstSchema(input, sourceSchema); err != nil {
		return nil, fmt.Errorf("input validation error: %w", err)
	}

	// 2. Transform
	output := make(map[string]interface{})
	for _, fm := range config.FieldMappings {
		rawVal, present := input[fm.SourceField]
		if !present {
			return nil, fmt.Errorf("field %q missing in input", fm.SourceField)
		}

		// Enum (string) -> bool mapping, e.g., "Active"/"Inactive" -> true/false
		if fm.ToBool && fm.ValueMap != nil {
			str, ok := rawVal.(string)
			if !ok {
				return nil, fmt.Errorf("expected string for field %q", fm.SourceField)
			}
			boolStr, exists := fm.ValueMap[str]
			if !exists {
				return nil, fmt.Errorf("unexpected value %q for field %q", str, fm.SourceField)
			}
			// Store as bool
			if boolStr == "true" {
				output[fm.TargetField] = true
			} else if boolStr == "false" {
				output[fm.TargetField] = false
			} else {
				return nil, fmt.Errorf("invalid bool mapping value %q for field %q", boolStr, fm.SourceField)
			}
			continue
		}

		// Bool -> Enum (string) mapping, e.g., true/false -> "Active"/"Inactive"
		if fm.FromBoolMap != nil {
			boolVal, ok := rawVal.(bool)
			if !ok {
				// Sometimes input JSON unmarshals "false"/"true" as strings - handle that
				if s, ok := rawVal.(string); ok {
					if s == "true" {
						boolVal = true
					} else if s == "false" {
						boolVal = false
					} else {
						return nil, fmt.Errorf("expected bool or string 'true'/'false' for field %q", fm.SourceField)
					}
				} else {
					return nil, fmt.Errorf("expected bool for field %q", fm.SourceField)
				}
			}
			strVal, exists := fm.FromBoolMap[boolVal]
			if !exists {
				return nil, fmt.Errorf("unexpected bool value for field %q", fm.SourceField)
			}
			output[fm.TargetField] = strVal
			continue
		}

		// Direct value mapping (optional)
		if fm.ValueMap != nil {
			str, ok := rawVal.(string)
			if !ok {
				return nil, fmt.Errorf("expected string for field %q", fm.SourceField)
			}
			mappedVal, exists := fm.ValueMap[str]
			if !exists {
				return nil, fmt.Errorf("unexpected mapping value %q for field %q", str, fm.SourceField)
			}
			output[fm.TargetField] = mappedVal
			continue
		}

		// String enum → int
		if fm.ToInt && fm.ValueStringToInt != nil {
			str, ok := rawVal.(string)
			if !ok {
				return nil, fmt.Errorf("expected string for field %q", fm.SourceField)
			}
			code, ok := fm.ValueStringToInt[str]
			if !ok {
				return nil, fmt.Errorf("unexpected value %q for field %q", str, fm.SourceField)
			}
			output[fm.TargetField] = code // _int_ value!
			continue
		}

		// Int → String enum
		if fm.ValueIntToString != nil {
			var intVal int
			switch v := rawVal.(type) {
			case float64:
				intVal = int(v)
			case int:
				intVal = v
			default:
				return nil, fmt.Errorf("expected int for field %q, got %T", fm.SourceField, rawVal)
			}
			strVal, ok := fm.ValueIntToString[intVal]
			if !ok {
				return nil, fmt.Errorf("unexpected int value %v for field %q", intVal, fm.SourceField)
			}
			output[fm.TargetField] = strVal
			continue
		}

		// Direct assignment
		output[fm.TargetField] = rawVal
	}

	// 3. Validate output against the destination schema
	if err := validateAgainstSchema(output, destSchema); err != nil {
		return nil, fmt.Errorf("output validation error: %w", err)
	}

	return output, nil
}
