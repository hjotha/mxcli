// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// MicroflowBackend provides microflow and nanoflow operations.
type MicroflowBackend interface {
	ListMicroflows() ([]*microflows.Microflow, error)
	GetMicroflow(id model.ID) (*microflows.Microflow, error)
	CreateMicroflow(mf *microflows.Microflow) error
	UpdateMicroflow(mf *microflows.Microflow) error
	DeleteMicroflow(id model.ID) error
	MoveMicroflow(mf *microflows.Microflow) error

	ListNanoflows() ([]*microflows.Nanoflow, error)
	GetNanoflow(id model.ID) (*microflows.Nanoflow, error)
	CreateNanoflow(nf *microflows.Nanoflow) error
	UpdateNanoflow(nf *microflows.Nanoflow) error
	DeleteNanoflow(id model.ID) error
	MoveNanoflow(nf *microflows.Nanoflow) error
}
