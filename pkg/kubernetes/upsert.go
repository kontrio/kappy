package kubernetes

import (
	"net/http"

	"github.com/apex/log"
	"k8s.io/apimachinery/pkg/api/errors"
)

type UpsertCommand struct {
	Create func() error
	Update func() error
}

func (u *UpsertCommand) Do() (err error) {
	err = u.Create()

	if err != nil {
		if apiErr, ok := err.(*errors.StatusError); ok {
			if apiErr.Status().Code == http.StatusConflict {
				err = u.Update()
				return
			}

			log.Debugf(apiErr.DebugError())

		}
		return
	}
	return
}
