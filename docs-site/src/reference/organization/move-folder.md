# MOVE FOLDER

## Synopsis

    MOVE FOLDER qualified_name TO FOLDER 'folder_path';
    MOVE FOLDER qualified_name TO FOLDER 'folder_path' IN module_name;
    MOVE FOLDER qualified_name TO module_name;

## Description

Moves a folder (and all its contents) to a different location: another folder within the same module, a folder in a different module, or a module root. The folder's contents (documents and sub-folders) move with it.

The source folder is identified using the standard qualified name syntax: `Module.FolderName`. For nested folder paths containing `/`, use double quotes: `Module."Parent/Child"`.

Target folders are created automatically if they don't exist.

## Parameters

**qualified_name**
: The source folder in `Module.FolderName` format. Use double quotes for paths with `/` or special characters.

**folder_path**
: The target folder path. Use `/` for nested folders.

**module_name**
: The target module name (for cross-module moves).

## Examples

### Move a folder into another folder

```sql
MOVE FOLDER MyModule.Resources TO FOLDER 'Archive';
```

### Move a nested folder to module root

```sql
MOVE FOLDER MyModule."Orders/Archive" TO MyModule;
```

### Move a folder to a different module

```sql
MOVE FOLDER MyModule.SharedWidgets TO CommonModule;
```

### Move a folder into a folder in another module

```sql
MOVE FOLDER MyModule.Templates TO FOLDER 'Shared/Templates' IN CommonModule;
```

## See Also

[CREATE FOLDER](create-folder.md), [DROP FOLDER](drop-folder.md), [MOVE](move.md)
