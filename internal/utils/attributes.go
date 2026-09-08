package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// Attribute value type identifiers shared by attribute value coercion across backends and the CLI.
const (
	AttributeValueTypeString  = "string"
	AttributeValueTypeInteger = "integer"
	AttributeValueTypeBoolean = "boolean"
)

// ConvertAttributeValue converts value to the type described by valueType (one of the AttributeValueType*
// constants), or to a slice if multiValued is true.
// TODO (Brynn Crowley): This is probably a good candidate for fuzz testing
//
//nolint:gocyclo
func ConvertAttributeValue(value any, valueType string, multiValued bool) (any, error) {
	if multiValued {
		if slice, ok := value.([]interface{}); ok {
			return slice, nil
		}

		if strSlice, ok := value.([]string); ok {
			result := make([]interface{}, len(strSlice))
			for i, s := range strSlice {
				result[i] = s
			}

			return result, nil
		}

		return []interface{}{value}, nil
	}

	switch valueType {
	case AttributeValueTypeBoolean:
		if boolVal, ok := value.(bool); ok {
			return boolVal, nil
		}

		if strVal, ok := value.(string); ok {
			switch strings.ToUpper(strVal) {
			case "TRUE", "T", "1", "YES", "Y":
				return true, nil
			case "FALSE", "F", "0", "NO", "N", "":
				return false, nil
			default:
				return nil, fmt.Errorf("invalid boolean value: %s", strVal)
			}
		}
	case AttributeValueTypeInteger:
		if intVal, ok := value.(int); ok {
			return int64(intVal), nil
		}

		if int64Val, ok := value.(int64); ok {
			return int64Val, nil
		}

		// JSON unmarshals numbers as float64.
		if floatVal, ok := value.(float64); ok {
			return int64(floatVal), nil
		}

		if strVal, ok := value.(string); ok {
			return strconv.ParseInt(strVal, 10, 64)
		}
	case AttributeValueTypeString, "":
		if strVal, ok := value.(string); ok {
			return strVal, nil
		}

		return fmt.Sprintf("%v", value), nil
	}

	return value, nil
}
