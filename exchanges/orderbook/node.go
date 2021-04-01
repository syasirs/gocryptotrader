package orderbook

import (
	"sync/atomic"
	"time"
)

const (
	cleanerActive    = 1
	cleanerNotActive = 0
)

var (
	defaultInterval  = time.Minute
	defaultAllowance = time.Second * 30
)

// node defines a linked list node for an orderbook item
type node struct {
	value Item
	next  *node
	prev  *node
	// Denotes time pushed to stack, this will influence cleanup routine when
	// there is a pause or minimal actions during period
	shelved time.Time
}

// stack defines a FILO list of reusable nodes
type stack struct {
	nodes []*node
	sema  uint32
	count int32
}

// newStack returns a ptr to a new stack instance, also starts the cleaning
// service
func newStack() *stack {
	s := &stack{}
	go s.cleaner()
	return s
}

// now defines a time which is now to ensure no other values get passed in
type now time.Time

// getNow returns the time at which it is called
func getNow() now {
	return now(time.Now())
}

// Push pushes a node pointer into the stack to be reused the time is passed in
// to allow for inlining which sets the time at which the node is theoretically
// pushed to a stack.
func (s *stack) Push(n *node, tn now) {
	if atomic.LoadUint32(&s.sema) == cleanerActive {
		// Cleaner is activated, for now we can derefence pointer
		n = nil
		return
	}
	// Adds a time when its placed back on to stack.
	n.shelved = time.Time(tn)
	n.next = nil
	n.prev = nil
	n.value = Item{}
	// Allows for resize when overflow TODO: rethink this
	s.nodes = append(s.nodes[:atomic.LoadInt32(&s.count)], n)
	atomic.AddInt32(&s.count, 1)
}

// Pop returns the last pointer off the stack and reduces the count and if empty
// will produce a lovely fresh node
func (s *stack) Pop() *node {
	if atomic.LoadUint32(&s.sema) == cleanerActive || atomic.LoadInt32(&s.count) == 0 {
		// Create an empty node when no nodes are in slice or when cleaning
		// service is running
		return &node{}
	}
	return s.nodes[atomic.AddInt32(&s.count, -1)]
}

// cleaner (POC) runs to the defaultTimer to clean excess nodes (nodes not being
// utilised) TODO: Couple time parameters to check for a reduction in activity.
// Add in counter per second function (?) so if there is a lot of activity don't
// inhibit stack performance.
func (s *stack) cleaner() {
	tt := time.NewTicker(defaultInterval)
sleeperino:
	for range tt.C {
		atomic.StoreUint32(&s.sema, cleanerActive)
		// As the old nodes are going to be left justified on this slice we
		// should just be able to shift the nodes that are still within time
		// allowance all the way to the left. Not going to resize capacity
		// because if it can get this big, it might as well stay this big.
		// TODO: Test and rethink if sizing is an issue
		nodesLen := atomic.LoadInt32(&s.count)
		for x := int32(0); x < nodesLen; x++ {
			if time.Since(s.nodes[x].shelved) > defaultAllowance {
				// Old node found continue
				continue
			}
			// First good node found, everything to the left of this on the
			// slice can be reassigned
			var counter int32
			for y := int32(0); y+x < nodesLen; y++ { // Go through good nodes
				// Reassign
				s.nodes[y] = s.nodes[y+x]
				// Add to the changed counter to remove from main
				// counter
				counter--
			}
			atomic.AddInt32(&s.count, counter)
			atomic.StoreUint32(&s.sema, cleanerNotActive)
			continue sleeperino
		}
		// Nodes are old, flush entirety.
		atomic.StoreInt32(&s.count, 0)
		atomic.StoreUint32(&s.sema, cleanerNotActive)
	}
}
