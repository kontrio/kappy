package taskrunner

import (
	"context"

	"github.com/google/uuid"
)

type taskRecord struct {
	id        string
	completed bool
	running   bool
	progress  int32
	context   context.Context
}

type Task struct {
	Description string
	PreRun      func(task *Task) error
	Run         func(task *Task) error

	dependencies map[string]*Task
	runRecord    *taskRecord
}

type Builder struct {
	Description string
	Async       bool
	PreRun      func(task *Task) error
	Run         func(task *Task) error
}

func NewTask(builder Builder) Task {
	return Task{
		Description: builder.Description,
		PreRun:      builder.PreRun,
		Run:         builder.Run,

		dependencies: make(map[string]*Task),
		runRecord: &taskRecord{
			id:        uuid.New().String(),
			completed: false,
			running:   false,
			progress:  0,
			context:   context.Background(),
		},
	}
}

func (t *Task) GetContext() context.Context {
	return t.runRecord.context
}

func (t *Task) isSelfCompleted() bool {
	return t.runRecord.completed
}

func (t *Task) setContext(ctx context.Context) {
	t.runRecord.context = ctx
}

func (t *Task) ExecuteSync(ctx context.Context) error {
	if t.isSelfCompleted() {
		return nil
	}

	t.setContext(ctx)

	if t.PreRun != nil {
		err := t.PreRun(t)
		if err != nil {
			return err
		}
	}

	dependencies := t.GetDependencies()

	for _, dependentTask := range dependencies {
		err := dependentTask.ExecuteSync(ctx)
		if err != nil {
			return err
		}
	}

	if t.Run != nil {
		err := t.Run(t)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Task) getUniqueId() string {
	return t.runRecord.id
}

func (t *Task) AddDependency(task *Task) *Task {
	t.dependencies[task.getUniqueId()] = task
	return t
}

func (t *Task) GetDependencies() []*Task {
	tasks := []*Task{}

	for _, task := range t.dependencies {
		tasks = append(tasks, task)
	}

	return tasks
}
