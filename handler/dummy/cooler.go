package dummy

import "time"

func (d *Dummy) Cooler() time.Duration {
	return 10 * time.Second
}
