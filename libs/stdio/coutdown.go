package stdio

import (
	"fmt"
	"time"
)

func CountdownWithBlink(duration time.Duration, blinkInterval time.Duration) {
	endTime := time.Now().Add(duration)

	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	for {
		remaining := time.Until(endTime)
		if remaining <= 0 {
			fmt.Print("\rOver!      ")
			break
		}

		seconds := int(remaining.Seconds()) + 1

		if time.Now().Second()%2 == 0 {
			fmt.Printf("\rLeft time: \033[7m%2ds\033[0m", seconds)
		} else {
			fmt.Printf("\rLeft time: %2ds", seconds)
		}

		time.Sleep(blinkInterval)
	}

	fmt.Println()
}
