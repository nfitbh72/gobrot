package threads

/**
 * the idea is to manage a pool of threads (goroutines)
 * a configurable numbver of parallel activities
 */

type TThreads struct {
	NumThreads uint16
	Routine    any
}

func (m *TThreads) Run() {

}
