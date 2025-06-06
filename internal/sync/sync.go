package sync

import (
	"karsai5/tw-caldav/internal/caldav"
	"karsai5/tw-caldav/internal/sync/task"
	"karsai5/tw-caldav/internal/tw"
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

func NewSyncProcess() (sp SyncProcess, err error) {
	local := tw.Taskwarrior{}
	remote, err := caldav.NewClient(viper.GetString("url"), viper.GetString("user"), viper.GetString("pass"))
	if err != nil {
		return sp, err
	}
	return SyncProcess{
		local:    local,
		remote:   *remote,
		synctime: time.Now(),
	}, err
}

type SyncProcess struct {
	local    tw.Taskwarrior
	remote   caldav.CalDavService
	synctime time.Time
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

	slog.Info("Tasks found", "locally", len(localTasks), "remotely", len(remoteTodos))

	taskGroups := processTasks(createMapOfLocalTasks(localTasks), createMapOfRemoteTasks(remoteTodos))

	if size := len(taskGroups.newRemoteTasks); size > 0 {
		slog.Info("Remote tasks to create", "num", size)
	}
	if size := len(taskGroups.newLocalTasks); size > 0 {
		slog.Info("Local tasks to create", "num", size)
	}
	if size := len(taskGroups.remoteTasksToDelete); size > 0 {
		slog.Info("Remote tasks to delete", "num", size)
	}
	if size := len(taskGroups.localTasksToDelete); size > 0 {
		slog.Info("Local tasks to delete", "num", size)
	}
	if size := len(taskGroups.tasksToUpdate); size > 0 {
		slog.Info("Tasks to update", "num", size)
	}

	for _, t := range taskGroups.newLocalTasks {
		slog.Info("Creating local task", "task", t.Description())
		uuid, err := sp.local.CreateTask(t, sp.synctime)
		if err != nil {
			slog.Error("Could not create local task", "err", err, "task", t)
			continue
		}

		remoteTaskUpdate := task.CreateShellTask(
			task.WithTask(t),
			task.WithSyncTime(sp.synctime),
			task.WithLocalId(uuid),
		)

		slog.Debug("Updating remote task", "task", remoteTaskUpdate.Task)
		err = t.Update(remoteTaskUpdate)
		if err != nil {
			slog.Error("Could not update remote task", "err", err, "task", t)
		}
	}

	for _, t := range taskGroups.localTasksToDelete {
		slog.Info("Deleting local task", "uuid", *t.LocalId(), "desc", t.Description())
		err := sp.local.RemoveTask(*t.LocalId())
		if err != nil {
			slog.Error("Error deleting task", "err", err)
		}
	}

	// TODO: create remote tasks

	// TODO: rm remote tasks

	// TODO: update tasks

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
				continue
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
