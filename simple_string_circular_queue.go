package tpcli

// SimpleStringCircularQueue is a simple circular buffer of strings
type SimpleStringCircularQueue struct {
	stringSlice             []string
	capacity                uint
	headIndex               uint
	indexOfNextInsert       uint
	indexOfLastSliceElement uint
	countOfItemsInQueue     uint
}

// NewSimpleStringCircularBuffer creates a simple circular buffer of strings, with
// the queue holding up to 'capacity' number of items.
func NewSimpleStringCircularBuffer(capacity uint) *SimpleStringCircularQueue {
	return &SimpleStringCircularQueue{
		stringSlice:             make([]string, capacity),
		headIndex:               0,
		indexOfNextInsert:       0,
		indexOfLastSliceElement: capacity - 1,
		countOfItemsInQueue:     0,
	}
}

// PutItemAtEnd places an item at the end of the circular queue
func (queue *SimpleStringCircularQueue) PutItemAtEnd(item string) {
	if queue.countOfItemsInQueue > 0 && queue.indexOfNextInsert == queue.headIndex {
		if queue.headIndex == queue.indexOfLastSliceElement {
			queue.headIndex = 0
		} else {
			queue.headIndex++
		}
	}

	queue.stringSlice[queue.indexOfNextInsert] = item

	if queue.indexOfNextInsert == queue.indexOfLastSliceElement {
		queue.indexOfNextInsert = 0
	} else {
		queue.indexOfNextInsert++
	}

	if queue.countOfItemsInQueue < uint(len(queue.stringSlice)) {
		queue.countOfItemsInQueue++
	}
}

// IsEmpty returns true if the queue has no items in it; false otherwise
func (queue *SimpleStringCircularQueue) IsEmpty() bool {
	return queue.countOfItemsInQueue == 0
}

// IsNotEmpty returns true if the queue has at least one item in it; false otherwise
func (queue *SimpleStringCircularQueue) IsNotEmpty() bool {
	return queue.countOfItemsInQueue != 0
}

// NumberOfItemsInTheQueue returns a count of the number of items in the queue
func (queue *SimpleStringCircularQueue) NumberOfItemsInTheQueue() uint {
	return queue.countOfItemsInQueue
}

// GetItemAtIndex retrieves the string at the specified index (0 is the first item)
func (queue *SimpleStringCircularQueue) GetItemAtIndex(index uint) (item string, thereIsAnItemAtThatIndex bool) {
	if queue.countOfItemsInQueue == 0 || index > queue.countOfItemsInQueue-1 {
		return "", false
	}

	sliceIndexOfItem := queue.headIndex + index
	if sliceIndexOfItem > queue.indexOfLastSliceElement {
		sliceIndexOfItem -= (queue.indexOfLastSliceElement + 1)
	}

	return queue.stringSlice[sliceIndexOfItem], true
}
