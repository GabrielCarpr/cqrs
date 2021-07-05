package bus

type Config = func(*Bus) error

// UseQueue provides an instantiated queue for the bus to use
func UseQueue(q Queue) Config {
	return func(b *Bus) error {
		b.queue = q
		return nil
	}
}
