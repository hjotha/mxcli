// SPDX-License-Identifier: Apache-2.0

package types

import "github.com/mendixlabs/mxcli/model"

// MPRVersion identifies the MPR file format version.
type MPRVersion int

const (
	MPRVersionV1 MPRVersion = 1 // single-file format
	MPRVersionV2 MPRVersion = 2 // mprcontents folder (Mendix 10.18+)
)

// ProjectVersion holds the parsed Mendix project version.
type ProjectVersion struct {
	ProductVersion string
	BuildVersion   string
	FormatVersion  int
	SchemaHash     string
	MajorVersion   int
	MinorVersion   int
	PatchVersion   int
}

// FolderInfo is a lightweight folder descriptor.
type FolderInfo struct {
	ID          model.ID
	ContainerID model.ID
	Name        string
}

// UnitInfo is a lightweight unit descriptor.
type UnitInfo struct {
	ID              model.ID
	ContainerID     model.ID
	ContainmentName string
	Type            string
}

// RenameHit records a single rename reference replacement.
type RenameHit struct {
	UnitID   string
	UnitType string
	Name     string
	Count    int
}

// RawUnit holds a unit's raw BSON contents.
type RawUnit struct {
	ID          model.ID
	ContainerID model.ID
	Type        string
	Contents    []byte
}

// RawUnitInfo holds a unit's raw contents with metadata.
type RawUnitInfo struct {
	ID            string
	QualifiedName string
	Type          string
	ModuleName    string
	Contents      []byte
}

// RawCustomWidgetType holds a custom widget's raw type/object data.
// RawType and RawObject are bson.D in sdk/mpr; here they are any to
// avoid a BSON driver dependency.
type RawCustomWidgetType struct {
	WidgetID   string
	RawType    any
	RawObject  any
	UnitID     string
	UnitName   string
	WidgetName string
}

// EntityMemberAccess describes access rights for a single entity member.
type EntityMemberAccess struct {
	AttributeRef   string
	AssociationRef string
	AccessRights   string
}

// EntityAccessRevocation describes which entity access to revoke.
type EntityAccessRevocation struct {
	RevokeCreate       bool
	RevokeDelete       bool
	RevokeReadMembers  []string
	RevokeWriteMembers []string
	RevokeReadAll      bool
	RevokeWriteAll     bool
}
