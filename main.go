package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type T interface{}

type BlockingQueue struct {
	m_DataArray  []T
	m_size       uint
	m_capacity   uint
	m_front      int
	m_back       int
	m_mutex      sync.Mutex
	queueIsFull  *sync.Cond
	queueIsEmpty *sync.Cond
}

func NewBlockingQueue(capacity uint) *BlockingQueue {
	instance := BlockingQueue{m_capacity: capacity, m_back: -1, m_mutex: sync.Mutex{}}
	instance.queueIsFull = sync.NewCond(&instance.m_mutex)
	instance.queueIsEmpty = sync.NewCond(&instance.m_mutex)
	instance.m_DataArray = make([]T, capacity)
	return &instance
}

func (instance *BlockingQueue) AtomicEnque(data T) {
	instance.m_mutex.Lock()
	for instance.m_size == instance.m_capacity {
		instance.queueIsFull.Wait()
	}
	instance.m_back = (instance.m_back + 1) % int(instance.m_capacity)
	instance.m_size += 1
	instance.m_DataArray[instance.m_back] = data
	fmt.Printf("Enqued %d\n", data)
	instance.m_mutex.Unlock()
	instance.queueIsEmpty.Signal()
}
func (instance *BlockingQueue) AtomicDeque() T {
	instance.m_mutex.Lock()
	for instance.m_size == 0 {
		instance.queueIsEmpty.Wait()
	}
	ReturnValue := instance.m_DataArray[instance.m_front]
	instance.m_front = (instance.m_front + 1) % int(instance.m_capacity)
	instance.m_size -= 1
	fmt.Printf("Dequed %d\n", ReturnValue)
	instance.m_mutex.Unlock()
	instance.queueIsFull.Signal()
	return ReturnValue
}
func (instance BlockingQueue) AtomicPeak() T {
	instance.m_mutex.Lock()
	Rv := instance.m_DataArray[instance.m_front]
	instance.m_mutex.Unlock()
	return Rv
}
func (instance BlockingQueue) Size() uint {
	instance.m_mutex.Lock()
	Rv := instance.m_size
	instance.m_mutex.Unlock()
	return Rv
}
func (instance BlockingQueue) Capacity() uint {
	instance.m_mutex.Lock()
	Rv := instance.m_capacity
	instance.m_mutex.Unlock()
	return Rv
}
func (instance BlockingQueue) IsEmpty() bool {
	instance.m_mutex.Lock()
	Rv := instance.m_size == 0
	instance.m_mutex.Unlock()
	return Rv
}
func (instance BlockingQueue) IsFull() bool {
	instance.m_mutex.Lock()
	Rv := instance.m_size == instance.m_capacity
	instance.m_mutex.Unlock()
	return Rv
}

var EnqueCounter atomic.Int64
var DequeCounter atomic.Int64

func addNumbers(queue *BlockingQueue, wg *sync.WaitGroup) {
	for i := range 100 {
		queue.AtomicEnque(i)
		EnqueCounter.Add(1)

	}
	defer wg.Done()
}
func deleteNumbers(queue *BlockingQueue, wg *sync.WaitGroup) {
	for {
		queue.AtomicDeque()
		DequeCounter.Add(1)
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	queue := NewBlockingQueue(5)
	var wg sync.WaitGroup
	wg.Add(8)
	go addNumbers(queue, &wg)
	go addNumbers(queue, &wg)
	go addNumbers(queue, &wg)
	go addNumbers(queue, &wg)
	go addNumbers(queue, &wg)
	go addNumbers(queue, &wg)
	go addNumbers(queue, &wg)
	go addNumbers(queue, &wg)
	go deleteNumbers(queue, &wg)
	go deleteNumbers(queue, &wg)

	wg.Wait()
	fmt.Printf("%d %d", EnqueCounter, DequeCounter)
}
