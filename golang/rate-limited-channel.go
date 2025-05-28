package rate_limited_channel

import "time"

// ToRateLimitedChannel takes an input channel and returns a new channel that
// will only send a value every delay duration. The value received will be the
// most recent value received on the input channel at the time that the
// duration expired.
func ToRateLimitedChannel[T any](input chan T, delay time.Duration) <-chan T {
	rateLimitedOutput := make(chan T)

	go func() {
		defer close(rateLimitedOutput)

		for {
			// Create timer for the specified delay duration
			timer := time.NewTimer(delay)

			var value T
			var receivedValue bool

			for {
				select {
				case newValue, ok := <-input:
					// Grab every value that comes in
					if !ok {
						// If the input channel is closed we'll return and close the output channel
						return
					}

					// Track our latest value and note that we've received a value
					value = newValue
					receivedValue = true
				case <-timer.C:
					// The timer has expired
					if !receivedValue {
						// If we haven't received a value we'll just send the next one
						newValue, ok := <-input

						if !ok {
							// If the input channel is closed we'll return and close the output channel
							return
						}

						// Use the value we just received
						value = newValue
					}

					// Send the value to the rate limited channel
					rateLimitedOutput <- value

					// Note that we're waiting for another value
					receivedValue = false

					// Reset the timer
					timer.Reset(delay)
				}
			}
		}
	}()

	return rateLimitedOutput
}
