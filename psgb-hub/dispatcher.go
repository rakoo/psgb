package main

type dispatcher struct {
	sh *subscribeHandler
	ph *publishHandler
}

func startDispatcher(sh *subscribeHandler, ph *publishHandler) {
	d := &dispatcher{sh, ph}

	go func() {
		for topic := range d.ph.newContent {
			d.sh.distributeToSubscribers(topic)
		}
	}()
}
