package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCallsUpdateIfCreateIsConflict(t *testing.T) {
	calledUpdated := false
	upsertCmd := UpsertCommand{
		Create: func() error {
			return errors.NewApplyConflict([]metav1.StatusCause{}, "")
		},
		Update: func() error {
			calledUpdate = true
			return nil
		},
	}

	assert.NotNil(t, upsertCmd.Do())
	assert.True(t, calledUpdate)
}

func TestDoesNotCallUpdateIfCreateIsNotErroring(t *testing.T) {
	calledUpdated := false
	upsertCmd := UpsertCommand{
		Create: func() error {
			return nil
		},
		Update: func() error {
			calledUpdate = true
			return nil
		},
	}

	assert.NotNil(t, upsertCmd.Do())
	assert.False(t, calledUpdate)
}
