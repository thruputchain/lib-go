package librpc

import "sync"

var (
	logLk    sync.Mutex
	logCalls bool
	logStats bool
)

func SetLogCalls(f bool) {
	logLk.Lock()
	defer logLk.Unlock()
	logCalls = f
}

func SetLogStats(f bool) {
	logLk.Lock()
	defer logLk.Unlock()
	logStats = f
}

func GetLogCalls() bool {
	logLk.Lock()
	defer logLk.Unlock()
	return logCalls
}

func GetLogStats() bool {
	logLk.Lock()
	defer logLk.Unlock()
	return logStats
}
