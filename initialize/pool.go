package initialize

import (
	"fmt"
	"sync"
	"time"
)
//任务
type Job interface {
	Do()//do something...
}

//worker 工人
type Worker struct {
	JobQueue chan Job   //任务队列
	Quit     chan bool //停止当前任务
}

//新建一个 worker 通道实例   新建一个工人
func NewWorker() Worker {
	return Worker{
		JobQueue: make(chan Job), //初始化工作队列为null
		Quit:     make(chan bool),
	}
}

/*
整个过程中 每个Worker(工人)都会被运行在一个协程中，
在整个WorkerPool(领导)中就会有num个可空闲的Worker(工人)，
当来一条数据的时候，领导就会小组中取一个空闲的Worker(工人)去执行该Job，
当工作池中没有可用的worker(工人)时，就会阻塞等待一个空闲的worker(工人)。
每读到一个通道参数 运行一个 worker
*/

func (w Worker) Run(wq chan chan Job){
	//这是一个独立的协程 循环读取通道内的数据，
  //保证 每读到一个通道参数就 去做这件事，没读到就阻塞
	go func() {
		for {
			wq <- w.JobQueue // 注册工作通道到线程池
			select {
			case job := w.JobQueue:
				job.Do() // 执行任务
			case <- w.Quit: 
			return // 停止任务
			}
		}
	}()
}

// Stop 方法控制 worker 停止监听工作请求
func (w Worker) Stop() {
	go func() {
			w.Quit <- true
	}()
}


// 线程池
type WorkerPool struct {
	workerlen int // 工人数量
	JobQueue chan Job // 任务队列
	WorkerQueue chan chan Job // 工人队列
}

func NewWorkerPool(workerlen int) *WorkerPool {
	return &WorkerPool{        
		workerlen:   workerlen,//开始建立 workerlen 个worker(工人)协程        
		JobQueue:    make(chan Job), //工作队列 通道        
		WorkerQueue: make(chan chan Job, workerlen), //最大通道参数设为 最大协程数 workerlen 工人的数量最大值    
	}
}

// 运行线程池
func (wp *WorkerPool) Run() {
	fmt.Println("开始运行线程池")
	for i:=0; i< wp.workerlen; i++ {
		worker := NewWorker() // 创建工人
		worker.Run(wp.WorkerQueue)
	}

	go func() {
		for {
			select {
			case job := <- wp.JobQueue: // 读取任务
			worker := <- wp.WorkerQueue // 获取一个空闲工人
			worker <- job // 任务分配给工人
			}
		}
	}()
}
