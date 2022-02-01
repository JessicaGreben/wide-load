package loader

import (
	"testing"
	"time"
)

func TestPipeline(t *testing.T) {
	load("https://google.com", 1, 1, 3*time.Second)
}
