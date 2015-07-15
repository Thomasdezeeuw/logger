package logger

import "errors"

// Combine combines multiple loggers into a single logger.
//
// Note: ShowDebug is enable by default and should be set on the individual
// loggers.
func Combine(name string, logs ...*Logger) (*Logger, error) {
	if len(logs) == 0 {
		return nil, errors.New("logger: Combine requires atleast one logger")
	}

	log, err := new(name, nil)
	if err != nil {
		return nil, err
	}
	log.ShowDebug = true

	go combinedLogWriter(log, logs)
	return log, nil
}

// Needs to be run in it's own goroutine, it blocks until log.logs is closed.
func combinedLogWriter(log *Logger, logs []*Logger) {
	// Pass on every message to the underlying loggers.
	for msg := range log.logs {
		for _, log := range logs {
			if msg.Level >= Debug || log.ShowDebug {
				log.logs <- msg
			}
		}
	}

	// Close all underlying loggers.
	errChan := make(chan error, len(logs))
	for _, log := range logs {
		go func(log *Logger) {
			errChan <- log.Close()
		}(log)
	}

	// Wait for all underlying loggers to respond.
	for i := len(logs); i > 0; i-- {
		if err := <-errChan; err != nil {
			log.Errors = append(log.Errors, err)
		}
	}

	// Add all underlying errors to the top one.
	for _, log2 := range logs {
		log.Errors = append(log.Errors, log2.Errors...)
	}

	log.closed <- struct{}{}
}
