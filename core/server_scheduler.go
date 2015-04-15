package core

import "time"

func (s *Server) SchedulePing(stopChan chan bool) {

	go func() {

		var err error
		var event *Event
		var latency time.Duration

		originalDelay := time.Second * time.Duration(s.Interval)
		delay := time.Second * time.Duration(s.Interval)

		for {

			latency, err = s.Ping()

			if err != nil {

				s.Log.Println(red, "ERROR: ", err, reset)

				// before sending an alert, pause 5 seconds & retry
				// prevent alerts from occaisional errors ('no such host' / 'i/o timeout') on cloud providers
				// todo: adjust sleep to fit with interval
				time.Sleep(5 * time.Second)
				_, rePingErr := s.Ping()
				if rePingErr != nil {

					// re-ping fails (confirms error)

					event = NewRedAlert(s, latency)
					s.StoreEvent(event)
					s.TriggerAlerts(event)

					s.IncrFailCount()
					if s.failCount > 0 {
						delay = time.Second * time.Duration(s.failCount*s.Interval)
					}

				} else {

					// re-ping succeeds (likely false positive)

					delay = originalDelay
					s.failCount = 0
				}

			} else {

				isRedalertRecovery := s.LastEvent != nil && s.LastEvent.isRedAlert()
				event = NewGreenAlert(s, latency)
				s.StoreEvent(event)
				if isRedalertRecovery {
					s.Log.Println(green, "RECOVERY: ", reset, s.Name)
					s.TriggerAlerts(event)
				}

				delay = originalDelay
				s.failCount = 0

			}

			select {
			case <-time.After(delay):
			case <-stopChan:
				return
			}
		}
	}()

}
