package commoncfg

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadConfigFile reads a JSON configuration file into the provided destination structure.
// An empty path results in a no-op.
func LoadConfigFile(path string, dst interface{}) error {
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("decode config file: %w", err)
	}

	return nil
}
