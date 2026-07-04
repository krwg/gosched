package engine

import (
	"encoding/json"
	"os"
)

func jsonMarshalImpl(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func writeFileImpl(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

func LoadResultJSON(path string) (*Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var res Result
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
