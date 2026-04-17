// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

// MappingBackend provides import/export mapping and JSON structure operations.
type MappingBackend interface {
	ListImportMappings() ([]*model.ImportMapping, error)
	GetImportMappingByQualifiedName(moduleName, name string) (*model.ImportMapping, error)
	CreateImportMapping(im *model.ImportMapping) error
	UpdateImportMapping(im *model.ImportMapping) error
	DeleteImportMapping(id model.ID) error
	MoveImportMapping(im *model.ImportMapping) error

	ListExportMappings() ([]*model.ExportMapping, error)
	GetExportMappingByQualifiedName(moduleName, name string) (*model.ExportMapping, error)
	CreateExportMapping(em *model.ExportMapping) error
	UpdateExportMapping(em *model.ExportMapping) error
	DeleteExportMapping(id model.ID) error
	MoveExportMapping(em *model.ExportMapping) error

	ListJsonStructures() ([]*mpr.JsonStructure, error)
	GetJsonStructureByQualifiedName(moduleName, name string) (*mpr.JsonStructure, error)
	CreateJsonStructure(js *mpr.JsonStructure) error
	DeleteJsonStructure(id string) error
}
