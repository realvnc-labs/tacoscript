package conv

import (
	"encoding/json"
	"fmt"
)

//ConvertSourceToJsonStrIfPossible converts any json capable types to str if not possible standard formatter is used
func ConvertSourceToJsonStrIfPossible(source interface{}) string {
	data, err := json.Marshal(source)
	if err != nil {
		return fmt.Sprintf("%+v", source)
	}

	return string(data)
}
