// Package storage provides CSV-based session persistence.
package storage

import (
	"encoding/csv"
	"fmt"
	"os"
)

// InitCsv creates the CSV file with a header row if it does not already exist.
func InitCsv(filename string) error {
	if _, err := os.Stat(filename); os.IsExist(err) {
		return nil
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error creating pomodoro.csv: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"topic", "start_time", "stop_time", "duration"}); err != nil {
		return fmt.Errorf("error writing to pomodoro.csv: %v", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("error writing to pomodoro.csv: %v", err)
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
// 		return fmt.Errorf("error writing to pomodoro.csv: %v", err)
// 	}

// 	writer.Flush()
// 	if err := writer.Error(); err != nil {
// 		return fmt.Errorf("error writing to pomodoro.csv: %v", err)
// 	}

// 	fmt.Println("Pomodoro added to csv")

// 	return nil
// }
