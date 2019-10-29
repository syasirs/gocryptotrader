package timedmutex

import (
	"testing"
	"time"
)

func BenchmarkTimedMutexTime(b *testing.B) {
	tm := NewTimedMutex(20 * time.Millisecond)
	for i := 0; i < b.N; i++ {
		tm.LockForDuration()
	}
}

func TestConsistencyOfPanicFreeUnlock(t *testing.T) {
	t.Parallel()
	duration := 20 * time.Millisecond
	tm := NewTimedMutex(duration)
	for i := 1; i <= 50; i++ {
		testUnlockTime := time.Duration(i) * time.Millisecond
		tm.LockForDuration()
		time.Sleep(testUnlockTime)
		tm.UnlockIfLocked()
	}
}

func TestUnlockAfterTimeout(t *testing.T) {
	t.Parallel()
	tm := NewTimedMutex(time.Second)
	tm.LockForDuration()
	time.Sleep(2 * time.Second)
	wasUnlocked := tm.UnlockIfLocked()
	if wasUnlocked {
		t.Error("Mutex should have been unlocked by timeout, not command")
	}
}

func TestUnlockBeforeTimeout(t *testing.T) {
	t.Parallel()
	tm := NewTimedMutex(2 * time.Second)
	tm.LockForDuration()
	time.Sleep(time.Second)
	wasUnlocked := tm.UnlockIfLocked()
	if !wasUnlocked {
		t.Error("Mutex should have been unlocked by command, not timeout")
	}
}

// TestUnlockAtSameTimeAsTimeout this test ensures
// that even if the timeout and the command occur at
// the same time, no panics occur. The result of the
// 'who' unlocking this doesn't matter, so long as
// the unlock occurs without this test panicking
func TestUnlockAtSameTimeAsTimeout(t *testing.T) {
	t.Parallel()
	duration := time.Second
	tm := NewTimedMutex(duration)
	tm.LockForDuration()
	time.Sleep(duration)
	tm.UnlockIfLocked()
}

func TestMultipleUnlocks(t *testing.T) {
	t.Parallel()
	tm := NewTimedMutex(10 * time.Second)
	tm.LockForDuration()
	wasUnlocked := tm.UnlockIfLocked()
	if !wasUnlocked {
		t.Error("Mutex should have been unlocked by command, not timeout")
	}
	wasUnlocked = tm.UnlockIfLocked()
	if wasUnlocked {
		t.Error("Mutex should have been already unlocked by command")
	}
	wasUnlocked = tm.UnlockIfLocked()
	if wasUnlocked {
		t.Error("Mutex should have been already unlocked by command")
	}
}

func TestJustWaitItOut(t *testing.T) {
	t.Parallel()
	tm := NewTimedMutex(1 * time.Second)
	tm.LockForDuration()
	time.Sleep(2 * time.Second)
}
