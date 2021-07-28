package bus

type Config = func(*Bus) error

// UseQueue provides an instantiated queue for the bus to use
func UseQueue(q Queue) Config {
	return func(b *Bus) error {
		b.queue = q
		return nil
	}
}

// UseEventStore provides an instantiated event store for the bus to use
func UseEventStore(s EventStore) Config {
	return func(b *Bus) error {
		b.eventStore = s
		return nil
	}
}
