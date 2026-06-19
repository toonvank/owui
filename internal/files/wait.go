package files

import (
	"fmt"
	"time"

	"github.com/toonvank/owui/internal/api"
)

const DefaultWaitTimeout = 5 * time.Minute

// WaitForProcessing polls until a file finishes processing or fails.
func WaitForProcessing(client *api.Client, id string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = DefaultWaitTimeout
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		st, err := client.FileStatus(id)
		if err != nil {
			return err
		}
		switch st.Status {
		case "completed":
			return nil
		case "failed":
			return fmt.Errorf("file processing failed: %s", st.Error)
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timed out waiting for file processing")
}