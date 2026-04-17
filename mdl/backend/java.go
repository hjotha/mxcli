// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/javaactions"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

// JavaBackend provides Java and JavaScript action operations.
type JavaBackend interface {
	ListJavaActions() ([]*mpr.JavaAction, error)
	ListJavaActionsFull() ([]*javaactions.JavaAction, error)
	ListJavaScriptActions() ([]*mpr.JavaScriptAction, error)
	ReadJavaActionByName(qualifiedName string) (*javaactions.JavaAction, error)
	ReadJavaScriptActionByName(qualifiedName string) (*mpr.JavaScriptAction, error)
	CreateJavaAction(ja *javaactions.JavaAction) error
	UpdateJavaAction(ja *javaactions.JavaAction) error
	DeleteJavaAction(id model.ID) error
	WriteJavaSourceFile(moduleName, actionName string, javaCode string, params []*javaactions.JavaActionParameter, returnType javaactions.CodeActionReturnType) error
	ReadJavaSourceFile(moduleName, actionName string) (string, error)
}
