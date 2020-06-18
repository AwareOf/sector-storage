package sectorstorage

import (
	"testing"
)

func TestTriggerSectorProcess(t *testing.T) {
	err := TriggerSectorProcess("DC", SealCommit2, 1, true)

	if err !=nil {
		t.Errorf("got error: %p", err)
	}
}

func TestTriggerNextPledge(t *testing.T) {
	err := TriggerNextPledge()
	if err !=nil {
		t.Errorf("got error: %p", err)
	}
}