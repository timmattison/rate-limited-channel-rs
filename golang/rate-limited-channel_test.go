package rate_limited_channel

import (
	"fmt"
	"testing"
	"time"
)

func TestRateLimitedChannelOncePerSecond(t *testing.T) {
	// Write values to the channel as fast as possible for 5.1 seconds
	testDuration := 5100 * time.Millisecond

	// Create a channel of integers for the input
	input := make(chan int)

	// Create a rate-limited channel that will output one value per second
	output := ToRateLimitedChannel(input, 1*time.Second)

	// Track how many values we've written
	numberOfValuesWritten := 0

	// Loop to write values until the test duration has elapsed
	go func() {
		endTime := time.Now().Add(testDuration)
		defer close(input)

		for {
			input <- numberOfValuesWritten
			numberOfValuesWritten++

			if time.Now().After(endTime) {
				break
			}
		}
	}()

	// Collect our values from the rate-limited output channel
	var values []int

	for value := range output {
		values = append(values, value)
	}

	result := fmt.Sprintf("Expected 5 values, received %d, wrote %d", len(values), numberOfValuesWritten)

	if len(values) != 5 {
		t.Fatal(result)
	}

	println(result)
}
