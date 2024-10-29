package main

import (
	"fmt"
	"sync"
	"time"
)

// WorkerPool структура для управления пулом воркеров
type WorkerPool struct {
	workerChan chan string    // Канал для задач
	workers    []*Worker      // Список воркеров
	mu         sync.Mutex     // Мьютекс для управления доступом
	stopChan   chan struct{}  // Канал для остановки воркеров
	wg         sync.WaitGroup // Группа для ожидания завершения всех воркеров
}

// Worker структура для одного воркера
type Worker struct {
	id   int
	stop chan struct{}
}

// NewWorkerPool создает новый воркер-пул
func NewWorkerPool() *WorkerPool {
	return &WorkerPool{
		workerChan: make(chan string),
		workers:    []*Worker{},
		stopChan:   make(chan struct{}),
	}
}

// AddWorker добавляет воркера в пул
func (wp *WorkerPool) AddWorker() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	worker := &Worker{
		id:   len(wp.workers) + 1,
		stop: make(chan struct{}),
	}
	wp.workers = append(wp.workers, worker)
	wp.wg.Add(1)
	go worker.Start(wp.workerChan, &wp.wg)
	fmt.Printf("Worker %d добавлен\n", worker.id)
}

// RemoveWorker удаляет воркера из пула
func (wp *WorkerPool) RemoveWorker() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if len(wp.workers) == 0 {
		fmt.Println("Нет воркеров для удаления")
		return
	}

	worker := wp.workers[len(wp.workers)-1]
	close(worker.stop)
	wp.workers = wp.workers[:len(wp.workers)-1]
	fmt.Printf("Worker %d удален\n", worker.id)
}

// Start запускает обработку данных в пуле
func (wp *WorkerPool) Start(data []string) {
	go func() {
		for _, d := range data {
			wp.workerChan <- d
			time.Sleep(500 * time.Millisecond) // Задержка для имитации поступления данных
		}
	}()
}

// Stop завершает работу пула
func (wp *WorkerPool) Stop() {
	close(wp.workerChan)
	wp.wg.Wait()
	fmt.Println("Все воркеры завершили работу")
}

// Start метод воркера для обработки данных
func (w *Worker) Start(dataChan <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case data, ok := <-dataChan:
			if !ok {
				return
			}
			fmt.Printf("Worker %d обрабатывает данные: %s\n", w.id, data)
		case <-w.stop:
			return
		}
	}
}

func main() {
	// Пример использования
	wp := NewWorkerPool()

	// Добавим несколько воркеров
	wp.AddWorker()
	wp.AddWorker()

	// Запустим обработку данных
	data := []string{"data1", "data2", "data3", "data4", "data5"}
	wp.Start(data)

	// Динамически добавим и удалим воркера
	time.Sleep(2 * time.Second)
	wp.AddWorker()
	time.Sleep(2 * time.Second)
	wp.RemoveWorker()

	// Ожидаем завершения всех воркеров
	wp.Stop()
}
