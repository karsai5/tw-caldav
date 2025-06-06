package sync

import (
	"fmt"
	"karsai5/tw-caldav/internal/caldav"
	"karsai5/tw-caldav/internal/sync/task"
	"karsai5/tw-caldav/internal/tw"
	"log/slog"

	"github.com/spf13/viper"
)

func NewSyncProcess() (sp SyncProcess, err error) {
	local := tw.Taskwarrior{}
	remote, err := caldav.NewClient(viper.GetString("url"), viper.GetString("user"), viper.GetString("pass"))
	if err != nil {
		return sp, err
	}
	return SyncProcess{
		local:  local,
		remote: *remote,
	}, err
}

type SyncProcess struct {
	local  tw.Taskwarrior
	remote caldav.CalDav
}

func (sp SyncProcess) Sync() error {

	localTasks, err := sp.local.GetAllTasks()
	if err != nil {
		return err
	}

	remoteTodos, err := sp.remote.GetAllTodos()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d tasks found locally\n", len(localTasks))
	fmt.Printf("%d tasks found remotely\n\n", len(remoteTodos))

	processedTasks := processTasks(createMapOfLocalTasks(localTasks), createMapOfRemoteTasks(remoteTodos))

	fmt.Printf("%d remote tasks to create\n", len(processedTasks.newRemoteTasks))
	if len(processedTasks.newRemoteTasks) > 0 {
		PrintTable(processedTasks.newRemoteTasks)
	}
	fmt.Printf("%d local tasks to create\n\n", len(processedTasks.newLocalTasks))
	if len(processedTasks.newLocalTasks) > 0 {
		PrintTable(processedTasks.newLocalTasks)
	}
	fmt.Printf("%d remote tasks to delete\n", len(processedTasks.remoteTasksToDelete))
	if len(processedTasks.remoteTasksToDelete) > 0 {
		PrintTable(processedTasks.remoteTasksToDelete)
	}
	fmt.Printf("%d local tasks to delete\n\n", len(processedTasks.localTasksToDelete))
	if len(processedTasks.localTasksToDelete) > 0 {
		PrintTable(processedTasks.localTasksToDelete)
	}
	fmt.Printf("%d tasks to update\n", len(processedTasks.tasksToUpdate))
	if len(processedTasks.tasksToUpdate) > 0 {
		arr := []task.Task{}
		for _, u := range processedTasks.tasksToUpdate {
			arr = append(arr, u.updatedTask)
		}
		PrintTable(arr)
	}
	return nil
}

type taskMapType map[string]task.Task

type taskToUpdate struct {
	localTask   task.Task
	remoteTask  task.Task
	updatedTask task.Task
}

type processedTasksReturn struct {
	newRemoteTasks      []task.Task
	newLocalTasks       []task.Task
	localTasksToDelete  []task.Task
	remoteTasksToDelete []task.Task
	tasksToUpdate       []taskToUpdate
}

func processTasks(localTaskMap taskMapType, remoteTaskMap taskMapType) processedTasksReturn {
	newRemoteTasks := []task.Task{}
	newLocalTasks := []task.Task{}
	localTasksToDelete := []task.Task{}
	remoteTasksToDelete := []task.Task{}
	tasksToUpdate := []taskToUpdate{}

	for _, t := range localTaskMap {
		if t.RemotePath() == nil {
			newRemoteTasks = append(newRemoteTasks, t)
			continue
		}

		if remoteTask, exists := remoteTaskMap[*t.RemotePath()]; exists {
			if t.LastSynced() == nil || remoteTask.LastSynced() == nil {
				slog.Error("Cant process task with no last synced time", "localtask", t, "remotetask", remoteTask)
			}
			if t.LastModified().After(*t.LastSynced()) || remoteTask.LastModified().After(*remoteTask.LastSynced()) {
				latestTask := t
				if remoteTask.LastModified().After(t.LastModified()) {
					latestTask = remoteTask
				}

				tasksToUpdate = append(tasksToUpdate, taskToUpdate{
					localTask:   t,
					remoteTask:  remoteTask,
					updatedTask: latestTask,
				})
			}
			continue
		} else {
			localTasksToDelete = append(localTasksToDelete, t)
			continue
		}

	}

	for _, t := range remoteTaskMap {
		if t.LocalId() == nil {
			newLocalTasks = append(newLocalTasks, t)
			continue
		}
		if _, exists := localTaskMap[*t.LocalId()]; !exists {
			remoteTasksToDelete = append(remoteTasksToDelete, t)
			continue
		}
	}
	return processedTasksReturn{
		newRemoteTasks:      newRemoteTasks,
		newLocalTasks:       newLocalTasks,
		localTasksToDelete:  localTasksToDelete,
		remoteTasksToDelete: remoteTasksToDelete,
		tasksToUpdate:       tasksToUpdate,
	}
}

func createMapOfLocalTasks(tasks []tw.Task) (taskMap taskMapType) {
	taskMap = make(taskMapType)
	for _, t := range tasks {
		taskMap[*t.LocalId()] = &t
	}
	return taskMap
}
func createMapOfRemoteTasks(tasks []caldav.Todo) (taskMap taskMapType) {
	taskMap = make(taskMapType)
	for _, t := range tasks {
		taskMap[*t.RemotePath()] = &t
	}
	return taskMap
}
