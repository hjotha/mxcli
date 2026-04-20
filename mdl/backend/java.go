// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/javaactions"
)

// JavaBackend provides Java and JavaScript action operations.
type JavaBackend interface {
	ListJavaActions() ([]*types.JavaAction, error)
	ListJavaActionsFull() ([]*javaactions.JavaAction, error)
	ListJavaScriptActions() ([]*types.JavaScriptAction, error)
	ReadJavaActionByName(qualifiedName string) (*javaactions.JavaAction, error)
	ReadJavaScriptActionByName(qualifiedName string) (*types.JavaScriptAction, error)
	CreateJavaAction(ja *javaactions.JavaAction) error
	UpdateJavaAction(ja *javaactions.JavaAction) error
	DeleteJavaAction(id model.ID) error
	WriteJavaSourceFile(moduleName, actionName string, javaCode string, params []*javaactions.JavaActionParameter, returnType javaactions.CodeActionReturnType) error
	ReadJavaSourceFile(moduleName, actionName string) (string, error)
}
