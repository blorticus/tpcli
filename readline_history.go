package tpcli

import (
	"github.com/blorticus/stringcque"
)

// ReadlineHistory stores items in an inverted stack and allows moving up and down that stack without removing the stack items.  One can move
// Up this inverted stack (toward the first entry) or Down this inverted stack (toward the last entry).  These are "Up" and "Down" because they
// map to arrow keys used in a cli readline.  The empty string is a not a valid entry and an effort to add an empty string is ignored.  Logically,
// the stack always has at least one item, and the last item is always the empty string.  If one moves Up from the first item, the first item is
// returned and the iterator doesn't move.  If one moves Down from the real last item, the empty string entry is returned.  If one attempts to move
// Down from this empty string, the empty string is returned and the iterator does not move.  If items are added, ResetIteration()
// should be called before trying to move Up or Down.  If this is not done, the results are undefined.
type ReadlineHistory struct {
	attachedQueue           *stringcque.SimpleStringCircularQueue
	indexOfLastItemReturned int // -1 is the implicit empty string entry at the bottom of the inverted stack
}

// NewReadlineHistory creates a ReadlineHistory which will contain up to maximumHistoryEntries items.  If the stack already has that number of
// entries and an item is added (via AddItem), then the least recently added entry is popped off the stack before the new item is added.
func NewReadlineHistory(maximumHistoryEntries uint) *ReadlineHistory {
	return &ReadlineHistory{
		attachedQueue:           stringcque.NewSimpleStringCircularBuffer(int(maximumHistoryEntries)),
		indexOfLastItemReturned: -1,
	}
}

// Up moves the iterator up the inverted stack, returning the item at the new iterator location.
func (history *ReadlineHistory) Up() string {
	if history.attachedQueue.IsEmpty() {
		return ""
	}

	if history.indexOfLastItemReturned > 0 {
		history.indexOfLastItemReturned--
	}

	value, thisIndexIsInQueue := history.attachedQueue.GetItemAtIndex(uint(history.indexOfLastItemReturned))

	if !thisIndexIsInQueue {
		return ""
	}

	return value
}

// Down moves the iterator down the inverted stack, returning the item at the new iterator location.
func (history *ReadlineHistory) Down() string {
	if history.attachedQueue.IsEmpty() {
		return ""
	}

	history.indexOfLastItemReturned++
	if history.indexOfLastItemReturned >= int(history.attachedQueue.NumberOfItemsInTheQueue()) {
		history.indexOfLastItemReturned = int(history.attachedQueue.NumberOfItemsInTheQueue())
		return ""
	}

	value, thisIndexIsInQueue := history.attachedQueue.GetItemAtIndex(uint(history.indexOfLastItemReturned))

	if !thisIndexIsInQueue {
		return ""
	}

	return value
}

// ResetIteration returns the iterator to the last entry in the inverted stack, which is always the implicit
// empty string entry.
func (history *ReadlineHistory) ResetIteration() {
	history.indexOfLastItemReturned = int(history.attachedQueue.NumberOfItemsInTheQueue())
}

// AddItem adds an item to the end of the inverted stack.
func (history *ReadlineHistory) AddItem(item string) {
	history.attachedQueue.PutItemAtEnd(item)
}
