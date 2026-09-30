package agent

import "fmt"

func waterReminder(w WaterTracker) string {
	return fmt.Sprintf("Time to drink water (%d/%d ml today)", w.ConsumedML, w.GoalML)
}

func breakReminder() string {
	return "Time for a break — stand up and stretch"
}
