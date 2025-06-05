package taskwarrior

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"
)

var TimeLayout = "20060102T150405Z"

type Task struct {
	Id          int        `json:"id"`
	Description string     `json:"description"`
	Due         *time.Time `json:"due"`
	Entry       string     `json:"entry"`
	Modified    time.Time  `json:"modified"`
	Project     string     `json:"project"`
	Status      string     `json:"status"`
	UUID        string     `json:"uuid"`
	Wait        string     `json:"wait"`
	Urgency     float32    `json:"urgency"`
	Tags        []string   `json:"tags"`
	CalDavId    string     `json:"caldavid"`
	Priority    string     `json:"priority"`
	RemotePath  string     `json:"remotepath"`
	LastSync    *time.Time `json:"lastsync"`
}

func (t *Task) UnmarshalJSON(data []byte) error {
	type Alias Task
	aux := &struct {
		Modified string `json:"modified"`
		Due      string `json:"due"`
		LastSync string `json:"lastsync"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	err := t.setModified(aux.Modified)
	if err != nil {
		return err
	}

	err = t.setDue(aux.Due)
	if err != nil {
		return err
	}

	err = t.setLastSync(aux.LastSync)
	if err != nil {
		return err
	}

	return nil
}

func (t *Task) setDue(str string) error {
	if str == "" {
		return nil
	}
	parsedTime, err := time.Parse(TimeLayout, str)
	if err != nil {
		return err
	}
	t.Due = &parsedTime
	return nil
}

func (t *Task) setLastSync(str string) error {
	if str == "" {
		return nil
	}
	parsedTime, err := time.Parse(TimeLayout, str)
	if err != nil {
		return err
	}
	t.LastSync = &parsedTime
	return nil
}

func (t *Task) setModified(str string) error {
	parsedTime, err := time.Parse(TimeLayout, str)
	if err != nil {
		return err
	}
	t.Modified = parsedTime
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

func Run(cmds ...string) (string, error) {
	out, err := exec.Command("task", cmds...).Output()

	if err != nil {
		return "", fmt.Errorf("while running task command: %w", err)
	}

	return string(out), nil
}

func List(filter string) (tasks []Task, err error) {
	cmdArgs := []string{}
	if filter != "" {
		cmdArgs = append(cmdArgs, filter)
	}
	cmdArgs = append(cmdArgs, "export")

	out, err := exec.Command("task", cmdArgs...).Output()
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
