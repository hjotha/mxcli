// SPDX-License-Identifier: Apache-2.0

// Package backend defines domain-specific interfaces that decouple the
// executor from concrete storage (e.g. .mpr files). Each interface
// groups related read/write operations by domain concept.
//
// Several method signatures currently reference types from sdk/mpr
// (e.g. NavigationDocument, FolderInfo, ImageCollection, JsonStructure,
// JavaAction, EntityMemberAccess, RenameHit). These should eventually be
// extracted into a shared types package to remove the mpr dependency.
package backend
