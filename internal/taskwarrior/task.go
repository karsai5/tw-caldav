package taskwarrior

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"
)

type Task struct {
	Id          int       `json:"id"`
	Description string    `json:"description"`
	Due         string    `json:"due"`
	Entry       string    `json:"entry"`
	Modified    time.Time `json:"modified"`
	Project     string    `json:"project"`
	Status      string    `json:"status"`
	UUID        string    `json:"uuid"`
	Wait        string    `json:"wait"`
	Urgency     float32   `json:"urgency"`
	Tags        []string  `json:"tags"`
	CalDavId    string    `json:"caldavid"`
}

func (t *Task) UnmarshalJSON(data []byte) error {
	type Alias Task
	aux := &struct {
		Modified string `json:"modified"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	parsed, err := time.Parse("20060102T150405Z", aux.Modified)
	if err != nil {
		return err
	}
	t.Modified = parsed
	return nil
}

func (t *Task) Append(str string) error {
	err := exec.Command("task", t.UUID, "append", str).Run()
	if err != nil {
		return fmt.Errorf("while appending to task: %w", err)
	}
	return nil
}

func Sync() {
	err := exec.Command("task", "sync").Run()
	if err != nil {
		log.Fatal(err)
	}
}

func List(filter string) (tasks []Task, err error) {
	out, err := exec.Command("task", filter, "export").Output()
	if err != nil {
		return tasks, fmt.Errorf("while running task command: %w", err)
	}

	err = json.Unmarshal(out, &tasks)
	if err != nil {
		return tasks, fmt.Errorf("while converting tasks from json: %w", err)
	}

	return tasks, nil
}

func Append(filter string, value string) error {
	return exec.Command("task", fmt.Sprintf("+PENDING and (%s)", filter), "append", value).Run()
}
