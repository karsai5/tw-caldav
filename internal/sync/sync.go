package sync

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/karsai5/tw-caldav/internal/caldav"
	"github.com/karsai5/tw-caldav/internal/sync/task"
	"github.com/karsai5/tw-caldav/internal/tw"

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
	local       tw.Taskwarrior
	remote      caldav.CalDavService
	synctime    time.Time
	Interactive bool
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

	taskGroups := processTasks(localTasks, remoteTodos)

	printTasks(taskGroups.newRemoteTasks, "Remote tasks to create")
	printTasks(taskGroups.newLocalTasks, "Local tasks to create")
	printTasks(taskGroups.remoteTasksToDelete, "Remote tasks to delete")
	printTasks(taskGroups.localTasksToDelete, "Local tasks to delete")

	if size := len(taskGroups.tasksToUpdate); size > 0 {
		slog.Info("Tasks to update", "num", size)
		for _, ttu := range taskGroups.tasksToUpdate {
			slog.Debug(ttu.updatedTask.Description(), "localId", *ttu.localTask.LocalId(), "remotePath", *ttu.remoteTask.RemotePath())
		}
	}

	sp.handleTasks(taskGroups.newLocalTasks, sp.handleLocalTaskCreate, "Would you like to create local tasks?", "Creating local tasks")
	sp.handleTasks(taskGroups.newRemoteTasks, sp.handleRemoteTaskCreate, "Would you like to create remote tasks?", "Creating remote tasks")

	sp.handleTasks(taskGroups.localTasksToDelete, sp.handleLocalTaskDelete, "Would you like to remove local tasks?", "Removing local tasks")
	sp.handleTasks(taskGroups.remoteTasksToDelete, sp.handleRemoteTaskDelete, "Would you like to remove remote tasks?", "Removing remote tasks")

	updateTasks := func() {
		for _, ttu := range taskGroups.tasksToUpdate {
			err := sp.handleTaskUpdate(ttu)
			if err != nil {
				slog.Error("Error updating task", "err", err)
			}
		}
	}

	if len(taskGroups.tasksToUpdate) > 0 {
		tasks := []task.Task{}
		for _, ttu := range taskGroups.tasksToUpdate {
			tasks = append(tasks, ttu.updatedTask)
		}
		if sp.Interactive {
			fmt.Println("Would you like to update tasks?")
			PrintTable(tasks)
			if yesNo() {
				updateTasks()
			}
		} else {
			slog.Info("Creating local tasks")
			updateTasks()
		}
	}

	return nil
}

func (sp SyncProcess) handleTasks(tasks []task.Task, handleFunc func(task.Task) error, interactionMsg string, logMsg string) {
	handletasks := func() {
		for _, t := range tasks {
			err := handleFunc(t)
			if err != nil {
				slog.Error("Error processing task", "err", err)
			}
		}
	}

	if len(tasks) == 0 {
		return
	}

	if sp.Interactive {
		fmt.Println(interactionMsg)
		PrintTable(tasks)
		if yesNo() {
			handletasks()
		}
	} else {
		slog.Info(logMsg)
		handletasks()
	}
}

func printTasks(tasks []task.Task, msg string) {
	if size := len(tasks); size > 0 {
		slog.Info(msg, "num", size)
		// PrintTable(tasks)
	}
}

func (sp SyncProcess) handleTaskUpdate(ttu taskToUpdate) error {
	slog.Info("Updating task", "task", ttu.updatedTask.Description(), "path", *ttu.updatedTask.RemotePath(), "uuid", *ttu.updatedTask.LocalId())

	updatedTask, err := ttu.remoteTask.Update(ttu.updatedTask)
	if err != nil {
		return fmt.Errorf("While updating remote task: %w", err)
	}

	_, err = ttu.localTask.Update(updatedTask)
	if err != nil {
		return fmt.Errorf("While updating local task: %w", err)
	}

	if equal, err := sp.AreTasksEqual(ttu.localTask, updatedTask); equal == false {
		if err != nil {
			return fmt.Errorf("While checking tasks were updated correctly: %w", err)
		}
		return fmt.Errorf("Tasks were not updated correctly, still appear to be unequal")
	}

	return nil
}

func (sp SyncProcess) AreTasksEqual(localTask task.Task, remoteTask task.Task) (bool, error) {
	currentLocalTask, err := sp.local.GetTask(*localTask.LocalId())
	if err != nil {
		return false, err
	}
	currentRemoteTask, err := sp.remote.GetTodo(remoteTask.Project(), *remoteTask.RemotePath())
	if err != nil {
		return false, err
	}

	equal := task.Equal(&currentLocalTask, &currentRemoteTask)
	if !equal {
		slog.Debug("local", "task", task.PrintTask(&currentLocalTask))
		slog.Debug("remot", "task", task.PrintTask(&currentRemoteTask))
	}

	return equal, nil
}

func (sp SyncProcess) handleRemoteTaskCreate(lt task.Task) error {
	slog.Info("Creating remote task", "task", lt.Description())
	finalPath, err := sp.remote.CreateNewTodo(lt)
	if err != nil {
		return err
	}
	slog.Info("Remote task created", "path", finalPath)

	localTaskUpdate := task.CreateShellTask(
		task.WithTask(lt),
		task.WithRemotePath(finalPath),
	)

	_, err = lt.Update(localTaskUpdate)
	if err != nil {
		return err
	}
	slog.Info("Local task updated", "uuid", lt.LocalId())

	return nil
}

func (sp SyncProcess) handleLocalTaskCreate(t task.Task) error {
	slog.Info("Creating local task", "task", t.Description())
	localTaskToAdd := task.CreateShellTask(task.WithTask(t))

	uuid, err := sp.local.AddTask(localTaskToAdd)
	if err != nil {
		return fmt.Errorf("While creating local task: %w", err)
	}

	remoteTaskUpdate := task.CreateShellTask(
		task.WithTask(t),
		task.WithLocalId(uuid),
	)

	slog.Debug("Updating remote task", "task", remoteTaskUpdate.Task)
	_, err = t.Update(remoteTaskUpdate)
	if err != nil {
		return fmt.Errorf("While updating remote task: %w", err)
	}
	return nil
}

func (sp SyncProcess) handleLocalTaskDelete(t task.Task) error {
	slog.Info("Deleting local task", "uuid", *t.LocalId(), "desc", t.Description())
	return t.Delete()
}

func (sp SyncProcess) handleRemoteTaskDelete(t task.Task) error {
	slog.Info("Deleting remote task", "uuid", *t.LocalId(), "desc", t.Description())
	return t.Delete()
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

func processTasks(localTasks []tw.Task, remoteTodos []caldav.Todo) processedTasksReturn {
	localTasksToDelete := []task.Task{}
	remoteTasksToDelete := []task.Task{}
	remoteTasksToCreate := []task.Task{}
	localTasksToCreate := []task.Task{}
	tasksToUpdate := []taskToUpdate{}

	localTaskMap := mapOfLocalTasks(localTasks)
	remoteTaskMap := mapOfRemoteTasks(remoteTodos)

	// Get remote tasks with no id, these need to be created locally
	for _, t := range remoteTodos {
		if t.LocalId() == nil {
			localTasksToCreate = append(localTasksToCreate, &t)
		}
	}

	// Get local tasks with no path, these need to be created remotely
	for _, t := range localTasks {
		if t.RemotePath() == nil {
			remoteTasksToCreate = append(remoteTasksToCreate, &t)
		}
	}

	// Remove any remote tasks that don't exist locally
	for uuid, t := range remoteTaskMap {
		if _, existsLocally := localTaskMap[uuid]; !existsLocally {
			remoteTasksToDelete = append(remoteTasksToDelete, t)
			delete(remoteTaskMap, uuid)
		}
	}

	// Remove any local tasks that have been synced but no longer exist remotely
	for uuid, t := range localTaskMap {
		if remotePath := t.RemotePath(); remotePath != nil {
			if _, existsRemotely := remoteTaskMap[uuid]; !existsRemotely {
				localTasksToDelete = append(localTasksToDelete, t)
				delete(localTaskMap, uuid)
			}
		}
	}

	// Find tasks with changes
	for uuid, t := range localTaskMap {
		if remoteTask, remoteTaskExists := remoteTaskMap[uuid]; remoteTaskExists {
			if !task.Equal(t, remoteTask) {
				slog.Debug("Tasks are not equal, update required")
				slog.Debug("local", "task", task.PrintTask(t))
				slog.Debug("remot", "task", task.PrintTask(remoteTask))
				tasksToUpdate = append(tasksToUpdate, taskToUpdate{
					localTask:   t,
					remoteTask:  remoteTask,
					updatedTask: getUpdateTask(t, remoteTask),
				})
			}
		}
	}

	return processedTasksReturn{
		newRemoteTasks:      remoteTasksToCreate,
		newLocalTasks:       localTasksToCreate,
		localTasksToDelete:  localTasksToDelete,
		remoteTasksToDelete: remoteTasksToDelete,
		tasksToUpdate:       tasksToUpdate,
	}
}

func getUpdateTask(a task.Task, b task.Task) task.Task {
	taskToUpdate := a
	if b.LastModified().After(a.LastModified()) {
		taskToUpdate = b
	}
	return taskToUpdate
}

func mapOfLocalTasks(tasks []tw.Task) taskMapType {
	taskMap := make(taskMapType)
	for _, t := range tasks {
		if t.LocalId() == nil {
			continue
		}
		taskMap[*t.LocalId()] = &t
	}
	return taskMap
}
func mapOfRemoteTasks(tasks []caldav.Todo) taskMapType {
	taskMap := make(taskMapType)
	for _, t := range tasks {
		if t.LocalId() == nil {
			continue
		}
		taskMap[*t.LocalId()] = &t
	}
	return taskMap
}

func createMapOfTasks(tasks []task.Task) (taskMap taskMapType) {
	taskMap = make(taskMapType)
	for _, t := range tasks {
		if t.LocalId() == nil {
			continue
		}
		taskMap[*t.LocalId()] = t
	}
	return taskMap
}
