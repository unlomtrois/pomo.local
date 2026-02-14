package storage

import (
	"encoding/csv"
	"fmt"
	"os"
)

func InitCsv(filename string) error {
	if _, err := os.Stat(filename); os.IsExist(err) {
		return nil
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Error creating pomodoro.csv: %v\n", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"title", "start_time", "stop_time", "duration"}); err != nil {
		return fmt.Errorf("Error writing to pomodoro.csv: %v\n", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("Error writing to pomodoro.csv: %v\n", err)
	}

	fmt.Println("pomodoro.csv initialized/exists")

	return nil
}

// func (p *Pomodoro) Save(filename string) error {
// 	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
// 	if err != nil {
// 		return fmt.Errorf("Error opening pomodoro.csv: %v\n", err)
// 	}
// 	defer file.Close()

// 	writer := csv.NewWriter(file)
// 	if err := writer.Write(p.Strings()); err != nil {
// 		return fmt.Errorf("Error writing to pomodoro.csv: %v\n", err)
// 	}

// 	writer.Flush()
// 	if err := writer.Error(); err != nil {
// 		return fmt.Errorf("Error writing to pomodoro.csv: %v\n", err)
// 	}

// 	fmt.Println("Pomodoro added to csv")

// 	return nil
// }
