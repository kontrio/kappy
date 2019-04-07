package taskrunner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldNotAllowDuplicateTaskDependencies(t *testing.T) {
	taskA := NewTask(Builder{
		Description: "testa",
	})

	taskB := NewTask(Builder{
		Description: "testb",
	})

	taskA.AddDependency(&taskB)
	taskA.AddDependency(&taskB)

	assert.Equal(t, 1, len(taskA.GetDependencies()))
}

func TestExecuteRunsAllDependentsBeforeSelf(t *testing.T) {
	ranTaskA := false
	ranTaskB := false
	ranTaskC := false

	taskA := NewTask(Builder{
		Description: "test",
		Run: func(task *Task) error {
			assert.False(t, ranTaskA)
			assert.True(t, ranTaskB)
			assert.True(t, ranTaskC)

			ranTaskA = true
			return nil
		},
	})

	taskB := NewTask(Builder{
		Description: "testb",
		Run: func(task *Task) error {
			assert.False(t, ranTaskA)
			assert.False(t, ranTaskB)
			assert.True(t, ranTaskC)

			ranTaskB = true
			return nil
		},
	})

	taskC := NewTask(Builder{
		Description: "testc",
		Run: func(task *Task) error {
			assert.False(t, ranTaskA)
			assert.False(t, ranTaskB)
			assert.False(t, ranTaskC)

			ranTaskC = true
			return nil
		},
	})

	taskB.AddDependency(&taskC)
	taskA.AddDependency(&taskB)

	err := taskA.ExecuteSync(context.Background())

	assert.Nil(t, err)
	assert.True(t, ranTaskA)
	assert.True(t, ranTaskB)
	assert.True(t, ranTaskC)
}
