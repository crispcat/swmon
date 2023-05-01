package main

import (
	"net"
	"sync"
)

type NetMethod func(task *NetTask, queue *NetTaskQueue, hostsModel *HostsModel)

type NetTask struct {
	ip     net.IP
	swargs SwmonNetworkArgs
	method NetMethod
}

type NetTaskQueue struct {
	tasks       chan NetTask
	tasksToWait *sync.WaitGroup
}

func (t *NetTaskQueue) Enqueue(task NetTask) {

	t.tasksToWait.Add(1)
	t.tasks <- task
}

func (t *NetTaskQueue) GetOne() NetTask {

	return <-t.tasks
}

func (t *NetTaskQueue) DoneOne() {

	t.tasksToWait.Done()
}

func (t *NetTaskQueue) WaitAllTasksCompletesAndClose() {

	t.tasksToWait.Wait()
	close(t.tasks)
}

func CreateTaskQueue(bufferSize uint32) *NetTaskQueue {

	tasks := NetTaskQueue{
		tasks:       make(chan NetTask, bufferSize),
		tasksToWait: new(sync.WaitGroup),
	}

	return &tasks
}

func NetWorker(queue *NetTaskQueue, hostsModel *HostsModel) {

	for task := range queue.tasks {
		task.method(&task, queue, hostsModel)
	}
}
