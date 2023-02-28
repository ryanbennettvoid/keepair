package node

import (
	"sync"

	"keepair/pkg/common"
)

type TransferOperation struct {
	SourceNode Node
	TargetNode Node
	Entry      common.Entry
}

func NewTransferOperation(entry common.Entry, source, target Node) TransferOperation {
	return TransferOperation{
		SourceNode: source,
		TargetNode: target,
		Entry:      entry,
	}
}

type Callback func(items []TransferOperation) error

type ITransferQueue interface {
	Push(item TransferOperation) error
	Flush() error
}

// TransferOperationsQueue is buffered queue of operations
// which flushes itself when the buffer size is reached. Should
// be explicitly flushed when finished using.
type TransferOperationsQueue struct {
	sync.RWMutex
	Items         []TransferOperation
	BufferSize    int
	FlushCallback Callback
}

func NewTransferOperationsQueueWithCallback(bufferSize int, flushCallback Callback) ITransferQueue {
	return &TransferOperationsQueue{
		Items:         make([]TransferOperation, 0),
		BufferSize:    bufferSize,
		FlushCallback: flushCallback,
	}
}

func (q *TransferOperationsQueue) Push(queueItem TransferOperation) error {

	q.Lock()
	q.Items = append(q.Items, queueItem)
	numOps := len(q.Items)
	q.Unlock()

	if numOps >= q.BufferSize {
		return q.Flush()
	}

	return nil
}

func (q *TransferOperationsQueue) Flush() error {
	q.Lock()
	defer q.Unlock()

	if err := q.FlushCallback(q.Items); err != nil {
		return err
	}

	q.Items = make([]TransferOperation, 0)

	return nil
}
