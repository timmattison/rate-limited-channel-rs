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

func TestRateLimitedChannelHighRate(t *testing.T) {
	// Write values to the channel as fast as possible for 0.5 seconds
	testDuration := 500 * time.Millisecond

	// Create a channel of integers for the input
	input := make(chan int)

	// Create a rate-limited channel that will output 1000 values per second
	output := ToRateLimitedChannel(input, 1*time.Millisecond)

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
	
	// Set a collection deadline slightly longer than test duration
	collectionDeadline := time.Now().Add(1 * time.Second)
	
	for {
		select {
		case value, ok := <-output:
			if !ok {
				// Channel closed
				goto DoneCollecting
			}
			values = append(values, value)
		case <-time.After(50 * time.Millisecond):
			// Check if we've reached the collection deadline
			if time.Now().After(collectionDeadline) {
				goto DoneCollecting
			}
		}
	}

DoneCollecting:
	result := fmt.Sprintf("High rate test: Received %d values, wrote %d", len(values), numberOfValuesWritten)
	println(result)

	// We should expect to receive approximately 500 values (1 per millisecond for 500ms)
	// But allowing for some system variation, we'll check for a reasonable minimum
	if len(values) < 400 {
		t.Fatal(fmt.Sprintf("Expected at least 400 values, received %d", len(values)))
	}
}

func TestRateLimitedChannelLowRate(t *testing.T) {
	// Write values to the channel as fast as possible for 12 seconds
	testDuration := 12000 * time.Millisecond

	// Create a channel of integers for the input
	input := make(chan int)

	// Create a rate-limited channel that will output one value per 10 seconds
	output := ToRateLimitedChannel(input, 10*time.Second)

	// Track how many values we've written
	numberOfValuesWritten := 0

	// Loop to write values until the test duration has elapsed
	go func() {
		endTime := time.Now().Add(testDuration)
		defer close(input)

		for {
			input <- numberOfValuesWritten
			numberOfValuesWritten++

			// Small sleep to avoid overwhelming the channel
			time.Sleep(100 * time.Millisecond)

			if time.Now().After(endTime) {
				break
			}
		}
	}()

	// Collect our values from the rate-limited output channel
	var values []int
	
	// Set a collection deadline slightly longer than test duration
	collectionDeadline := time.Now().Add(12500 * time.Millisecond)
	
	for {
		select {
		case value, ok := <-output:
			if !ok {
				// Channel closed
				goto DoneCollecting
			}
			values = append(values, value)
		case <-time.After(250 * time.Millisecond):
			// Check if we've reached the collection deadline
			if time.Now().After(collectionDeadline) {
				goto DoneCollecting
			}
		}
	}

DoneCollecting:
	result := fmt.Sprintf("Low rate test: Received %d values, wrote %d", len(values), numberOfValuesWritten)
	println(result)

	// With a 10-second rate limit and a 12-second test, we should see at least 1 value
	if len(values) < 1 {
		t.Fatal(fmt.Sprintf("Expected at least 1 value, received %d", len(values)))
	}
}
