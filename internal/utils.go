package internal

import (
	"encoding/json"
	"fmt"
	"github.com/xeipuuv/gojsonschema"
)

func intPtr(v int) *int {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

func validate(schema string, data map[string]any) error {
	b, err := schemas.ReadFile(schema)
	if err != nil {
		return err
	}

	if data == nil {
		data = make(map[string]any)
	}

	res, err := gojsonschema.Validate(gojsonschema.NewBytesLoader(b), gojsonschema.NewGoLoader(data))
	if err != nil {
		return err
	}

	if !res.Valid() {
		return fmt.Errorf("invalid data: %s", res.Errors())
	}

	return nil
}

func loadSchemaNode(filename string, dst any) error {
	body, err := schemas.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dst)
}
