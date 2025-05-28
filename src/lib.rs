use std::time::Duration;
use tokio::sync::mpsc::{self, Receiver, Sender};
use tokio::time::{sleep, Instant};

/// Creates a rate-limited channel from an input channel.
///
/// This function takes an input channel receiver and returns a new channel
/// that will only send a value every specified delay duration. The value sent will be the
/// most recent value received from the input channel at the time the duration expired.
///
/// # Arguments
///
/// * `input` - The receiver part of an input channel
/// * `delay` - The minimum duration between each value sent on the output channel
///
/// # Returns
///
/// A receiver that outputs values from the input channel at the specified rate
pub fn to_rate_limited_channel<T: Clone + Send + 'static>(
    input: Receiver<T>,
    delay: Duration,
) -> Receiver<T> {
    let (output_tx, output_rx) = mpsc::channel::<T>(100);
    
    tokio::spawn(async move {
        rate_limit_worker(input, output_tx, delay).await;
    });
    
    output_rx
}

/// Worker function that processes the input channel and sends to the output channel
/// at the specified rate.
async fn rate_limit_worker<T: Clone + Send>(
    mut input: Receiver<T>,
    output: Sender<T>, 
    delay: Duration
) {
    let mut last_send_time = Instant::now() - delay; // Allow immediate first send
    
    while let Some(mut value) = input.recv().await {
        // Keep accepting values while waiting for the delay to elapse
        let mut received_additional_value = true;
        
        while received_additional_value {
            // Check if we should send now (enough time has passed)
            let now = Instant::now();
            let time_since_last_send = now.duration_since(last_send_time);
            
            if time_since_last_send >= delay {
                // Delay elapsed, send the current value
                if let Err(_) = output.send(value.clone()).await {
                    // Output channel closed, exit worker
                    return;
                }
                
                last_send_time = now;
                break;
            }
            
            // Calculate remaining time to wait
            let wait_time = delay.saturating_sub(time_since_last_send);
            
            // Wait for either new value or timeout
            received_additional_value = false;
            tokio::select! {
                // Either wait for a new message...
                new_value = input.recv() => {
                    match new_value {
                        Some(v) => {
                            // Got a new value, use it instead
                            value = v;
                            received_additional_value = true;
                        },
                        None => {
                            // Input channel closed
                            return;
                        }
                    }
                },
                // ...or wait for the delay time
                _ = sleep(wait_time) => {
                    // Time's up, ready to send in the next loop iteration
                }
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tokio::sync::mpsc;
    use tokio::time::{sleep, Duration};
    
    #[tokio::test]
    async fn test_rate_limited_channel_once_per_second() {
        const TEST_DURATION: Duration = Duration::from_millis(5100);
        const RATE_LIMIT: Duration = Duration::from_secs(1);
        
        // Create input channel
        let (input_tx, input_rx) = mpsc::channel::<i32>(1000);
        
        // Create rate-limited output channel
        let mut output_rx = to_rate_limited_channel(input_rx, RATE_LIMIT);
        
        // Track how many values we've written
        let mut number_of_values_written = 0;
        
        // Start the writer task
        let writer_handle = tokio::spawn(async move {
            let end_time = Instant::now() + TEST_DURATION;
            
            while Instant::now() < end_time {
                if let Err(_) = input_tx.send(number_of_values_written).await {
                    break;
                }
                number_of_values_written += 1;
                
                // Small sleep to avoid overwhelming the channel
                sleep(Duration::from_micros(10)).await;
            }
            
            number_of_values_written
        });
        
        // Collect values from the rate-limited output channel with timeout
        let mut values = Vec::new();
        let collection_timeout = Duration::from_millis(5500); // Just a bit longer than test duration
        let collection_deadline = Instant::now() + collection_timeout;
        
        while Instant::now() < collection_deadline {
            tokio::select! {
                Some(value) = output_rx.recv() => {
                    values.push(value);
                    
                    // Break if we have collected exactly 5 values
                    if values.len() == 5 {
                        break;
                    }
                }
                _ = sleep(collection_timeout) => {
                    break;
                }
            }
        }
        
        // Wait for the writer task to complete
        let number_of_values_written = writer_handle.await.unwrap();
        
        println!("Expected 5 values, received {}, wrote {}", values.len(), number_of_values_written);
        
        assert_eq!(values.len(), 5);
    }
}
