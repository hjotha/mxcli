// SPDX-License-Identifier: Apache-2.0

package executor

// Registration stubs — each function will register handlers for its domain
// during the handler migration phase. For now they are no-ops so the
// registry compiles and the existing type-switch dispatch continues to work.

func registerConnectionHandlers(_ *Registry)         {}
func registerModuleHandlers(_ *Registry)             {}
func registerEnumerationHandlers(_ *Registry)        {}
func registerConstantHandlers(_ *Registry)           {}
func registerDatabaseConnectionHandlers(_ *Registry) {}
func registerEntityHandlers(_ *Registry)             {}
func registerAssociationHandlers(_ *Registry)        {}
func registerMicroflowHandlers(_ *Registry)          {}
func registerPageHandlers(_ *Registry)               {}
func registerSecurityHandlers(_ *Registry)           {}
func registerNavigationHandlers(_ *Registry)         {}
func registerImageHandlers(_ *Registry)              {}
func registerWorkflowHandlers(_ *Registry)           {}
func registerBusinessEventHandlers(_ *Registry)      {}
func registerSettingsHandlers(_ *Registry)           {}
func registerODataHandlers(_ *Registry)              {}
func registerJSONStructureHandlers(_ *Registry)      {}
func registerMappingHandlers(_ *Registry)            {}
func registerRESTHandlers(_ *Registry)               {}
func registerDataTransformerHandlers(_ *Registry)    {}
func registerQueryHandlers(_ *Registry)              {}
func registerStylingHandlers(_ *Registry)            {}
func registerRepositoryHandlers(_ *Registry)         {}
func registerSessionHandlers(_ *Registry)            {}
func registerLintHandlers(_ *Registry)               {}
func registerAlterPageHandlers(_ *Registry)          {}
func registerFragmentHandlers(_ *Registry)           {}
func registerSQLHandlers(_ *Registry)                {}
func registerImportHandlers(_ *Registry)             {}
