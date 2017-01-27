package pulse

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestMtrImplWithTimeout(t *testing.T) {
	req := &MtrRequest{
		Target: "www.example.com.",
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	resp := MtrImpl(ctx, req)
	cancel()
	if !strings.Contains(resp.Err, "context deadline exceeded") {
		t.Errorf("unexpected error: %s", resp.Err)
	}
	if resp.ErrEnglish != "Test was cancelled because agent was unresponsible for 50 seconds during test execution. This may indicate agent is malfunctioning; please inform maintainers." {
		t.Errorf("unexpected ErrEnglish: %s", resp.ErrEnglish)
	}
}
