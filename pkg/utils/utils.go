// pkg/utils/utils.go

package utils

import (
	"encoding/json"

	"gorm.io/datatypes"
)

// StringSliceToJSON converts a slice of strings to datatypes.JSON.
func StringSliceToJSON(slice []string) datatypes.JSON {
	data, err := json.Marshal(slice)
	if err != nil {
		return datatypes.JSON("[]")
	}
	return datatypes.JSON(data)
}
