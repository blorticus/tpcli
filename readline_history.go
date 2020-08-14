package tpcli

import (
	"github.com/blorticus/stringcque"
)

type readlineHistory struct {
	attachedQueue           *stringcque.SimpleStringCircularQueue
	indexOfLastItemReturned uint
	iterationHasStarted     bool
}

func newReadlineHistory(maximumHistoryEntries uint) *readlineHistory {
	return &readlineHistory{
		attachedQueue:           stringcque.NewSimpleStringCircularBuffer(int(maximumHistoryEntries)),
		indexOfLastItemReturned: 0,
		iterationHasStarted:     false,
	}
}

func (history *readlineHistory) Up() (historyItem string, wasNotYetAtTopOfList bool) {
	if history.attachedQueue.IsNotEmpty() {
		if history.iterationHasStarted {
			if history.iteratorIsNotAtStartOfHistoryList() {
				v, _ := history.attachedQueue.GetItemAtIndex(history.indexOfLastItemReturned - 1)
				history.indexOfLastItemReturned--
				return v, true
			}
		} else {
			history.iterationHasStarted = true
			v, _ := history.attachedQueue.GetItemAtIndex(history.attachedQueue.NumberOfItemsInTheQueue() - 1)
			history.indexOfLastItemReturned = history.attachedQueue.NumberOfItemsInTheQueue() - 1
			return v, true
		}
	}

	return "", false
}

func (history *readlineHistory) Down() (historyItem string, wasNotYetAtBottomOfList bool) {
	if history.attachedQueue.IsNotEmpty() {
		if history.iterationHasStarted {
			if history.iteratorIsNotAtEndOfHistoryList() {
				v, _ := history.attachedQueue.GetItemAtIndex(history.indexOfLastItemReturned + 1)
				history.indexOfLastItemReturned++
				return v, true
			}
		}
	}

	return "", false
}

func (history *readlineHistory) iteratorIsNotAtEndOfHistoryList() bool {
	return history.attachedQueue.NumberOfItemsInTheQueue() > history.indexOfLastItemReturned+1
}

func (history *readlineHistory) iteratorIsNotAtStartOfHistoryList() bool {
	return history.indexOfLastItemReturned != 0
}

func (history *readlineHistory) ResetIteration() {
	history.indexOfLastItemReturned = 0
	history.iterationHasStarted = false
}

func (history *readlineHistory) AddItem(item string) {
	history.attachedQueue.PutItemAtEnd(item)
}
