package ircon

import (
	"time"
)

type backoff struct {
	mindur, maxdur time.Duration
	last           time.Time
	delay          time.Duration
}

func newBackoff(minsec, maxsec int) *backoff {
	return &backoff{
		last:   time.Now(),
		delay:  time.Second * time.Duration(minsec),
		mindur: time.Second * time.Duration(minsec),
		maxdur: time.Second * time.Duration(maxsec),
	}
}

func (cd *backoff) Now() {
	cd.last = time.Now()
}

func (cd *backoff) Delay() time.Duration {
	delta := time.Since(cd.last)
	if delta >= cd.delay {
		// Time between last try as longer than the delay, so reduce
		cd.delay = cd.delay / 2
		if cd.delay < cd.mindur {
			cd.delay = cd.mindur
		}
	} else {
		cd.delay += cd.mindur
		if cd.delay > cd.maxdur {
			cd.delay = cd.maxdur
		}
	}
	interval := cd.delay - delta
	if interval < 0 {
		interval = 0
	}
	return interval
}
