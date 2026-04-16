// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/mpr"
	"github.com/mendixlabs/mxcli/sdk/pages"
	"github.com/mendixlabs/mxcli/sdk/widgets"
)

func (pb *pageBuilder) buildDataViewV3(w *ast.WidgetV3) (*pages.DataView, error) {
	dv := &pages.DataView{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$DataView",
			},
			Name: w.Name,
		},
	}

	// Handle DataSource
	if ds := w.GetDataSource(); ds != nil {
		dataSource, entityName, err := pb.buildDataSourceV3(ds)
		if err != nil {
			return nil, fmt.Errorf("failed to build datasource: %w", err)
		}
		dv.DataSource = dataSource

		// Save and restore entity context so nested DataViews work correctly
		oldContext := pb.entityContext
		pb.entityContext = entityName
		defer func() { pb.entityContext = oldContext }()

		// Register the widget name with its entity so template params like $dvOrder.Attr
		// can be resolved to Entity.Attr
		if w.Name != "" && entityName != "" {
			pb.paramEntityNames[w.Name] = entityName
		}
	}

	// Build child widgets, separating FOOTER widgets into FooterWidgets
	for _, child := range w.Children {
		// Check if this is a FOOTER widget - its children go to FooterWidgets
		if child.Type == "FOOTER" {
			dv.ShowFooter = true
			for _, fw := range child.Children {
				widget, err := pb.buildWidgetV3(fw)
				if err != nil {
					return nil, err
				}
				dv.FooterWidgets = append(dv.FooterWidgets, widget)
			}
			continue
		}
		childWidget, err := pb.buildWidgetV3(child)
		if err != nil {
			return nil, err
		}
		dv.Widgets = append(dv.Widgets, childWidget)
	}

	// Also build footer widgets from Properties (legacy support)
	if footerWidgets, ok := w.Properties["Footer"].([]*ast.WidgetV3); ok {
		dv.ShowFooter = true
		for _, fw := range footerWidgets {
			widget, err := pb.buildWidgetV3(fw)
			if err != nil {
				return nil, err
			}
			dv.FooterWidgets = append(dv.FooterWidgets, widget)
		}
	}

	if err := pb.registerWidgetName(w.Name, dv.ID); err != nil {
		return nil, err
	}

	return dv, nil
}

func (pb *pageBuilder) buildDataGridV3(w *ast.WidgetV3) (*pages.CustomWidget, error) {
	// Build DataGrid2 as a CustomWidget (pluggable widget) like V2 does.
	// The built-in DataGrid (Forms$DataGrid) has serialization issues.
	widgetID := model.ID(mpr.GenerateID())

	// Load embedded template (required for pluggable widgets to work)
	embeddedType, embeddedObject, embeddedIDs, embeddedObjectTypeID, err := widgets.GetTemplateFullBSON(pages.WidgetIDDataGrid2, mpr.GenerateID, pb.reader.Path())
	if err != nil {
		return nil, fmt.Errorf("failed to load DataGrid2 template: %w", err)
	}
	if embeddedType == nil || embeddedObject == nil {
		return nil, fmt.Errorf("DataGrid2 template not found")
	}

	// Convert widget IDs to pages.PropertyTypeIDEntry format
	propertyTypeIDs := convertPropertyTypeIDs(embeddedIDs)

	// Build datasource from V3 DataSource property
	var datasource pages.DataSource
	if ds := w.GetDataSource(); ds != nil {
		dataSource, entityName, err := pb.buildDataSourceV3(ds)
		if err != nil {
			return nil, fmt.Errorf("failed to build datasource: %w", err)
		}
		datasource = dataSource

		// Save and restore entity context so nested containers work correctly
		oldContext := pb.entityContext
		pb.entityContext = entityName
		defer func() { pb.entityContext = oldContext }()
	}

	// Extract column definitions and CONTROLBAR widgets from children
	var columns []ast.DataGridColumnDef
	var headerWidgets []bson.D
	for _, child := range w.Children {
		switch strings.ToUpper(child.Type) {
		case "COLUMN":
			attr := child.GetAttribute()
			// Sugar: when no explicit Attribute: property is given, fall back to
			// the column's name. This lets `COLUMN Sku (Caption: 'SKU')` work
			// without repeating `Attribute: Sku`. Skip for custom-content columns
			// (those with a body of child widgets), which don't bind to an attribute.
			if attr == "" && child.Name != "" && len(child.Children) == 0 {
				attr = child.Name
			}
			col := ast.DataGridColumnDef{
				Attribute:  attr,
				Caption:    child.GetCaption(),
				ChildrenV3: child.Children, // Child widgets for custom content columns
				Properties: child.Properties,
			}
			columns = append(columns, col)
		case "CONTROLBAR":
			// Build CONTROLBAR widgets as BSON for the filtersPlaceholder property
			for _, controlBarChild := range child.Children {
				widgetBSON, err := pb.buildWidgetV3ToBSON(controlBarChild)
				if err != nil {
					return nil, fmt.Errorf("failed to build controlbar widget: %w", err)
				}
				if widgetBSON != nil {
					headerWidgets = append(headerWidgets, widgetBSON)
				}
			}
		}
	}

	// Update the template object with datasource, columns, and header widgets
	var updatedObject bson.D
	if len(columns) > 0 || len(headerWidgets) > 0 {
		// Use full update that replaces columns and/or header widgets
		updatedObject = pb.updateDataGrid2Object(embeddedObject, propertyTypeIDs, datasource, columns, headerWidgets)
	} else {
		// No columns or header widgets defined, use template columns
		updatedObject = pb.cloneDataGrid2ObjectWithDatasourceOnly(embeddedObject, propertyTypeIDs, datasource)
	}

	// Apply paging properties from AST if specified
	updatedObject = pb.applyDataGridPagingProps(updatedObject, propertyTypeIDs, w)

	// Apply selection mode if specified
	if selection := w.GetSelection(); selection != "" {
		updatedObject = pb.applyDataGridSelectionProp(updatedObject, propertyTypeIDs, selection)
	}

	grid := &pages.CustomWidget{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       widgetID,
				TypeName: "CustomWidgets$CustomWidget",
			},
			Name: w.Name,
		},
		Editable:          "Always",
		RawType:           embeddedType,
		RawObject:         updatedObject,
		PropertyTypeIDMap: propertyTypeIDs,
		ObjectTypeID:      embeddedObjectTypeID,
	}

	if err := pb.registerWidgetName(w.Name, grid.ID); err != nil {
		return nil, err
	}

	return grid, nil
}

func (pb *pageBuilder) buildDataGridColumnV3(w *ast.WidgetV3) (*pages.DataGridColumn, error) {
	col := &pages.DataGridColumn{
		BaseElement: model.BaseElement{
			ID:       model.ID(mpr.GenerateID()),
			TypeName: "Forms$DataGridColumn",
		},
		Name:     w.Name,
		Editable: true,
	}

	// Get attribute from Attribute property
	if attr := w.GetAttribute(); attr != "" {
		col.AttributePath = pb.resolveAttributePath(attr)
	}

	// Get caption
	if caption := w.GetCaption(); caption != "" {
		col.Caption = &model.Text{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Texts$Text",
			},
			Translations: map[string]string{"en_US": caption},
		}
	}

	return col, nil
}

func (pb *pageBuilder) buildListViewV3(w *ast.WidgetV3) (*pages.ListView, error) {
	lv := &pages.ListView{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$ListView",
			},
			Name: w.Name,
		},
		PageSize: 20,
	}

	// Handle DataSource
	if ds := w.GetDataSource(); ds != nil {
		dataSource, entityName, err := pb.buildDataSourceV3(ds)
		if err != nil {
			return nil, fmt.Errorf("failed to build datasource: %w", err)
		}
		lv.DataSource = dataSource

		// Save and restore entity context so nested containers work correctly
		oldContext := pb.entityContext
		pb.entityContext = entityName
		defer func() { pb.entityContext = oldContext }()

		// Register widget name with entity for SELECTION datasource lookup
		if w.Name != "" && entityName != "" {
			pb.paramEntityNames[w.Name] = entityName
		}
	}

	// Register widget scope for SELECTION references
	if err := pb.registerWidgetName(w.Name, lv.ID); err != nil {
		return nil, err
	}

	// Build template widgets
	for _, child := range w.Children {
		widget, err := pb.buildWidgetV3(child)
		if err != nil {
			return nil, err
		}
		lv.Widgets = append(lv.Widgets, widget)
	}

	return lv, nil
}

func (pb *pageBuilder) buildTextBoxV3(w *ast.WidgetV3) (*pages.TextBox, error) {
	tb := &pages.TextBox{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$TextBox",
			},
			Name: w.Name,
		},
	}

	// Handle Attribute (attribute path)
	if attr := w.GetAttribute(); attr != "" {
		tb.AttributePath = pb.resolveAttributePath(attr)
	}

	// Handle Label
	if label := w.GetLabel(); label != "" {
		tb.Label = label
	}

	if err := pb.registerWidgetName(w.Name, tb.ID); err != nil {
		return nil, err
	}

	return tb, nil
}

func (pb *pageBuilder) buildTextAreaV3(w *ast.WidgetV3) (*pages.TextArea, error) {
	ta := &pages.TextArea{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$TextArea",
			},
			Name: w.Name,
		},
	}

	// Handle Attribute
	if attr := w.GetAttribute(); attr != "" {
		ta.AttributePath = pb.resolveAttributePath(attr)
	}

	// Handle Label
	if label := w.GetLabel(); label != "" {
		ta.Label = label
	}

	if err := pb.registerWidgetName(w.Name, ta.ID); err != nil {
		return nil, err
	}

	return ta, nil
}

func (pb *pageBuilder) buildDatePickerV3(w *ast.WidgetV3) (*pages.DatePicker, error) {
	dp := &pages.DatePicker{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$DatePicker",
			},
			Name: w.Name,
		},
	}

	// Handle Attribute
	if attr := w.GetAttribute(); attr != "" {
		dp.AttributePath = pb.resolveAttributePath(attr)
	}

	// Handle Label
	if label := w.GetLabel(); label != "" {
		dp.Label = label
	}

	if err := pb.registerWidgetName(w.Name, dp.ID); err != nil {
		return nil, err
	}

	return dp, nil
}

func (pb *pageBuilder) buildDropdownV3(w *ast.WidgetV3) (*pages.DropDown, error) {
	dd := &pages.DropDown{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$DropDown",
			},
			Name: w.Name,
		},
	}

	// Handle Attribute
	if attr := w.GetAttribute(); attr != "" {
		dd.AttributePath = pb.resolveAttributePath(attr)
	}

	// Handle Label
	if label := w.GetLabel(); label != "" {
		dd.Label = label
	}

	if err := pb.registerWidgetName(w.Name, dd.ID); err != nil {
		return nil, err
	}

	return dd, nil
}

func (pb *pageBuilder) buildCheckBoxV3(w *ast.WidgetV3) (*pages.CheckBox, error) {
	cb := &pages.CheckBox{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$CheckBox",
			},
			Name: w.Name,
		},
	}

	// Handle Attribute
	if attr := w.GetAttribute(); attr != "" {
		cb.AttributePath = pb.resolveAttributePath(attr)
	}

	// Handle Label
	if label := w.GetLabel(); label != "" {
		cb.Label = label
	}

	if err := pb.registerWidgetName(w.Name, cb.ID); err != nil {
		return nil, err
	}

	return cb, nil
}

// buildRadioButtonsV3 creates RadioButtons from V3 syntax.
func (pb *pageBuilder) buildRadioButtonsV3(w *ast.WidgetV3) (*pages.RadioButtons, error) {
	rb := &pages.RadioButtons{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$RadioButtonGroup",
			},
			Name: w.Name,
		},
		Label: w.GetLabel(),
	}

	// Get attribute path from Attribute property
	if attr := w.GetAttribute(); attr != "" {
		rb.AttributePath = pb.resolveAttributePath(attr)
	}

	if err := pb.registerWidgetName(w.Name, rb.ID); err != nil {
		return nil, err
	}

	return rb, nil
}

func (pb *pageBuilder) buildTextWidgetV3(w *ast.WidgetV3) (*pages.Text, error) {
	st := &pages.Text{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$Text",
			},
			Name: w.Name,
		},
		RenderMode: pages.TextRenderModeText,
	}

	// Handle Content
	if content := w.GetContent(); content != "" {
		st.Caption = &model.Text{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Texts$Text",
			},
			Translations: map[string]string{"en_US": content},
		}
	}

	// Handle RenderMode
	if rm := w.GetRenderMode(); rm != "" {
		st.RenderMode = pages.TextRenderMode(rm)
	}

	if err := pb.registerWidgetName(w.Name, st.ID); err != nil {
		return nil, err
	}

	return st, nil
}

func (pb *pageBuilder) buildDynamicTextV3(w *ast.WidgetV3) (*pages.DynamicText, error) {
	dt := &pages.DynamicText{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$DynamicText",
			},
			Name: w.Name,
		},
		RenderMode: pages.TextRenderModeText,
	}

	// Handle RenderMode
	if rm := w.GetRenderMode(); rm != "" {
		dt.RenderMode = pages.TextRenderMode(rm)
	}

	// Handle Content
	content := w.GetContent()
	explicitParams := w.GetContentParams()

	// Check if Content is an attribute reference AND no explicit params provided
	// If so, auto-generate template {1} and add the attribute as a parameter
	// Examples:
	//   Content: $widget.Name            -> auto-generate {1} with $widget.Name as param
	//   Content: Entity.Attribute        -> auto-generate {1} with Entity.Attribute as param
	//   Content: SomeStaticText          -> literal string, no params (no dot, no $)
	//   Content: 'Name: {1}', ContentParams: [Name] -> use explicit template and params
	var autoGeneratedParams []string
	if content != "" && explicitParams == nil {
		// Only auto-generate for:
		// - Variable references: $var or $widget.Attr (starts with $)
		// - Entity paths: Entity.Attribute (identifier.identifier pattern, not version numbers like "1.0")
		// Simple identifiers without dots are treated as static text
		isEntityPath := false
		if strings.Contains(content, ".") && !strings.HasPrefix(content, "$") {
			// Check if it looks like Entity.Attribute (letter followed by word chars, dot, letter followed by word chars)
			// This avoids matching strings like "Version 1.0" or "Dashboard - V2.1"
			isEntityPath = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*\.[A-Za-z_][A-Za-z0-9_]*$`).MatchString(content)
		}
		if strings.HasPrefix(content, "$") || isEntityPath {
			autoGeneratedParams = append(autoGeneratedParams, content)
			content = "{1}"
		}
	}

	if content == "" {
		content = "{1}"
	}

	dt.Content = &pages.ClientTemplate{
		BaseElement: model.BaseElement{
			ID:       model.ID(mpr.GenerateID()),
			TypeName: "Forms$ClientTemplate",
		},
		Template: &model.Text{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Texts$Text",
			},
			Translations: map[string]string{"en_US": content},
		},
	}

	// Add auto-generated parameters first
	for _, attrRef := range autoGeneratedParams {
		param := &pages.ClientTemplateParameter{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$ClientTemplateParameter",
			},
		}
		pb.resolveTemplateAttributePathFull(attrRef, param)
		dt.Content.Parameters = append(dt.Content.Parameters, param)
	}

	// Handle explicit ContentParams
	if explicitParams != nil {
		for _, p := range explicitParams {
			param := &pages.ClientTemplateParameter{
				BaseElement: model.BaseElement{
					ID:       model.ID(mpr.GenerateID()),
					TypeName: "Forms$ClientTemplateParameter",
				},
			}
			// Check if it's an attribute reference or literal
			if strVal, ok := p.Value.(string); ok {
				if strings.HasPrefix(strVal, "'") || strings.HasPrefix(strVal, "\"") {
					// Already a quoted string literal - use as-is
					param.Expression = strVal
				} else if strings.HasPrefix(strVal, "$") || strings.Contains(strVal, ".") {
					// Attribute reference - resolve widget references to entity paths
					pb.resolveTemplateAttributePathFull(strVal, param)
				} else {
					// Unquoted literal value - assume attribute in current context
					pb.resolveTemplateAttributePathFull(strVal, param)
				}
			}
			dt.Content.Parameters = append(dt.Content.Parameters, param)
		}
	}

	if err := pb.registerWidgetName(w.Name, dt.ID); err != nil {
		return nil, err
	}

	return dt, nil
}

func (pb *pageBuilder) buildTitleV3(w *ast.WidgetV3) (*pages.Title, error) {
	title := &pages.Title{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$Title",
			},
			Name: w.Name,
		},
	}

	// Set caption from Content property
	content := w.GetContent()
	if content != "" {
		title.Caption = &model.Text{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Texts$Text",
			},
			Translations: map[string]string{"en_US": content},
		}
	}

	if err := pb.registerWidgetName(w.Name, title.ID); err != nil {
		return nil, err
	}

	return title, nil
}

func (pb *pageBuilder) buildButtonV3(w *ast.WidgetV3) (*pages.ActionButton, error) {
	btn := &pages.ActionButton{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$ActionButton",
			},
			Name: w.Name,
		},
		ButtonStyle: pages.ButtonStyleDefault,
	}

	// Handle Caption
	if caption := w.GetCaption(); caption != "" {
		btn.CaptionTemplate = &pages.ClientTemplate{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$ClientTemplate",
			},
			Template: &model.Text{
				BaseElement: model.BaseElement{
					ID:       model.ID(mpr.GenerateID()),
					TypeName: "Texts$Text",
				},
				Translations: map[string]string{"en_US": caption},
			},
		}

		// Handle CaptionParams (template parameters like {1}, {2})
		if params := w.GetCaptionParams(); params != nil {
			for _, p := range params {
				param := &pages.ClientTemplateParameter{
					BaseElement: model.BaseElement{
						ID:       model.ID(mpr.GenerateID()),
						TypeName: "Forms$ClientTemplateParameter",
					},
				}
				// Check if it's an attribute reference or literal
				if strVal, ok := p.Value.(string); ok {
					if strings.HasPrefix(strVal, "'") || strings.HasPrefix(strVal, "\"") {
						// Already a quoted string literal - use as-is
						param.Expression = strVal
					} else if strings.HasPrefix(strVal, "$") || strings.Contains(strVal, ".") {
						// Attribute reference - resolve widget references to entity paths
						param.AttributeRef = pb.resolveTemplateAttributePath(strVal)
					} else {
						// Unquoted literal value - wrap in quotes for expression
						param.Expression = "'" + strVal + "'"
					}
				}
				btn.CaptionTemplate.Parameters = append(btn.CaptionTemplate.Parameters, param)
			}
		}
	}

	// Handle ButtonStyle
	if style := w.GetButtonStyle(); style != "" {
		btn.ButtonStyle = pages.ButtonStyle(style)
	}

	// Handle Action
	if action := w.GetAction(); action != nil {
		act, err := pb.buildClientActionV3(action)
		if err != nil {
			return nil, fmt.Errorf("failed to build action: %w", err)
		}
		btn.Action = act
	}

	if err := pb.registerWidgetName(w.Name, btn.ID); err != nil {
		return nil, err
	}

	return btn, nil
}

// buildNavigationListV3 creates a NavigationList widget from V3 syntax.
func (pb *pageBuilder) buildNavigationListV3(w *ast.WidgetV3) (*pages.NavigationList, error) {
	navList := &pages.NavigationList{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$NavigationList",
			},
			Name: w.Name,
		},
	}

	// Build items from children (ITEM widgets)
	for _, child := range w.Children {
		if strings.ToUpper(child.Type) == "ITEM" {
			item, err := pb.buildNavigationListItemV3(child)
			if err != nil {
				return nil, err
			}
			navList.Items = append(navList.Items, item)
		}
	}

	if err := pb.registerWidgetName(w.Name, navList.ID); err != nil {
		return nil, err
	}

	return navList, nil
}

// buildNavigationListItemV3 creates a NavigationListItem from V3 syntax.
func (pb *pageBuilder) buildNavigationListItemV3(w *ast.WidgetV3) (*pages.NavigationListItem, error) {
	if w.Name == "" {
		return nil, fmt.Errorf("ITEM inside NAVIGATIONLIST requires a name")
	}

	item := &pages.NavigationListItem{
		BaseElement: model.BaseElement{
			ID:       model.ID(mpr.GenerateID()),
			TypeName: "Forms$NavigationListItem",
		},
		Name: w.Name,
	}

	if err := pb.registerWidgetName(w.Name, item.ID); err != nil {
		return nil, err
	}

	// Set caption from Caption property
	if caption := w.GetCaption(); caption != "" {
		item.Caption = &model.Text{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Texts$Text",
			},
			Translations: map[string]string{"en_US": caption},
		}
	}

	// Handle Action property
	if action := w.GetAction(); action != nil {
		clientAction, err := pb.buildClientActionV3(action)
		if err != nil {
			return nil, err
		}
		item.Action = clientAction
	}

	// Build child widgets
	for _, child := range w.Children {
		childWidget, err := pb.buildWidgetV3(child)
		if err != nil {
			return nil, err
		}
		item.Widgets = append(item.Widgets, childWidget)
	}

	return item, nil
}

// buildSnippetCallV3 creates a SnippetCallWidget from V3 syntax.
func (pb *pageBuilder) buildSnippetCallV3(w *ast.WidgetV3) (*pages.SnippetCallWidget, error) {
	sc := &pages.SnippetCallWidget{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$SnippetCallWidget",
			},
			Name: w.Name,
		},
	}

	// Handle Snippet property - resolve snippet and store both ID and name
	if snippetName := w.GetSnippet(); snippetName != "" {
		snippetID, err := pb.resolveSnippetRef(snippetName)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve snippet %s: %w", snippetName, err)
		}
		sc.SnippetID = snippetID
		sc.SnippetName = snippetName // Store qualified name for BY_NAME_REFERENCE serialization
	}

	if err := pb.registerWidgetName(w.Name, sc.ID); err != nil {
		return nil, err
	}

	return sc, nil
}

// buildTemplateV3 creates a Container to hold template content.
func (pb *pageBuilder) buildTemplateV3(w *ast.WidgetV3) (*pages.Container, error) {
	container := &pages.Container{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$DivContainer",
			},
			Name: w.Name,
		},
	}

	// Build children
	for _, child := range w.Children {
		childWidget, err := pb.buildWidgetV3(child)
		if err != nil {
			return nil, err
		}
		container.Widgets = append(container.Widgets, childWidget)
	}

	return container, nil
}

// buildFilterV3 creates a Container to hold filter widgets.
func (pb *pageBuilder) buildFilterV3(w *ast.WidgetV3) (*pages.Container, error) {
	container := &pages.Container{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$DivContainer",
			},
			Name: w.Name,
		},
	}

	// Build children (filter widgets)
	for _, child := range w.Children {
		childWidget, err := pb.buildWidgetV3(child)
		if err != nil {
			return nil, err
		}
		container.Widgets = append(container.Widgets, childWidget)
	}

	return container, nil
}

func (pb *pageBuilder) buildStaticImageV3(w *ast.WidgetV3) (*pages.StaticImage, error) {
	img := &pages.StaticImage{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$StaticImageViewer",
			},
			Name: w.Name,
		},
		Responsive: true,
	}

	if width := w.GetIntProp("Width"); width > 0 {
		img.Width = width
	}
	if height := w.GetIntProp("Height"); height > 0 {
		img.Height = height
	}

	if err := pb.registerWidgetName(w.Name, img.ID); err != nil {
		return nil, err
	}

	return img, nil
}

func (pb *pageBuilder) buildDynamicImageV3(w *ast.WidgetV3) (*pages.DynamicImage, error) {
	img := &pages.DynamicImage{
		BaseWidget: pages.BaseWidget{
			BaseElement: model.BaseElement{
				ID:       model.ID(mpr.GenerateID()),
				TypeName: "Forms$ImageViewer",
			},
			Name: w.Name,
		},
		Responsive: true,
	}

	if width := w.GetIntProp("Width"); width > 0 {
		img.Width = width
	}
	if height := w.GetIntProp("Height"); height > 0 {
		img.Height = height
	}

	if err := pb.registerWidgetName(w.Name, img.ID); err != nil {
		return nil, err
	}

	return img, nil
}

// dataGridPagingPropMap maps PascalCase MDL property names to camelCase widget property keys.
var dataGridPagingPropMap = map[string]string{
	"PageSize":          "pageSize",
	"Pagination":        "pagination",
	"PagingPosition":    "pagingPosition",
	"ShowPagingButtons": "showPagingButtons",
	// "ShowNumberOfRows" is defined in DataGrid2 type but not yet fully supported;
	// setting it to a non-default value causes CE0463 "widget definition changed".
}

// applyDataGridPagingProps applies paging properties from the AST to the DataGrid2 BSON object.
// It iterates through the object's Properties array, matching TypePointers to known paging
// property keys, and replaces PrimitiveValue when a corresponding AST property is set.
func (pb *pageBuilder) applyDataGridPagingProps(obj bson.D, propertyTypeIDs map[string]pages.PropertyTypeIDEntry, w *ast.WidgetV3) bson.D {
	// Collect overrides: camelCase key -> string value
	overrides := make(map[string]string)
	for mdlKey, widgetKey := range dataGridPagingPropMap {
		if v := w.GetStringProp(mdlKey); v != "" {
			overrides[widgetKey] = v
		} else if iv := w.GetIntProp(mdlKey); iv > 0 {
			overrides[widgetKey] = fmt.Sprintf("%d", iv)
		} else if bv, ok := w.Properties[mdlKey]; ok {
			if boolVal, isBool := bv.(bool); isBool {
				if boolVal {
					overrides[widgetKey] = "yes"
				} else {
					overrides[widgetKey] = "no"
				}
			}
		}
	}
	if len(overrides) == 0 {
		return obj
	}

	// Build reverse map: TypePointer ID -> widget property key
	typePointerToKey := make(map[string]string)
	for widgetKey, entry := range propertyTypeIDs {
		typePointerToKey[entry.PropertyTypeID] = widgetKey
	}

	// Walk the object and replace properties that have overrides
	result := make(bson.D, 0, len(obj))
	for _, elem := range obj {
		if elem.Key == "Properties" {
			if propsArr, ok := elem.Value.(bson.A); ok {
				updatedProps := bson.A{propsArr[0]} // Keep version marker
				for _, propVal := range propsArr[1:] {
					propMap, ok := propVal.(bson.D)
					if !ok {
						updatedProps = append(updatedProps, propVal)
						continue
					}
					tp := pb.getTypePointerFromProperty(propMap)
					widgetKey := typePointerToKey[tp]
					if newVal, hasOverride := overrides[widgetKey]; hasOverride {
						updatedProps = append(updatedProps, pb.clonePropertyWithPrimitiveValue(propMap, newVal))
					} else {
						updatedProps = append(updatedProps, propMap)
					}
				}
				result = append(result, bson.E{Key: "Properties", Value: updatedProps})
			} else {
				result = append(result, elem)
			}
		} else {
			result = append(result, elem)
		}
	}
	return result
}

// applyDataGridSelectionProp applies the Selection mode to a DataGrid2 object.
// DataGrid2 uses the same "itemSelection" property key as Gallery.
func (pb *pageBuilder) applyDataGridSelectionProp(obj bson.D, propertyTypeIDs map[string]pages.PropertyTypeIDEntry, selectionMode string) bson.D {
	itemSelectionEntry, ok := propertyTypeIDs["itemSelection"]
	if !ok {
		return obj
	}

	result := make(bson.D, 0, len(obj))
	for _, elem := range obj {
		if elem.Key == "Properties" {
			if propsArr, ok := elem.Value.(bson.A); ok {
				updatedProps := bson.A{propsArr[0]} // Keep version marker
				for _, propVal := range propsArr[1:] {
					propMap, ok := propVal.(bson.D)
					if !ok {
						updatedProps = append(updatedProps, propVal)
						continue
					}
					tp := pb.getTypePointerFromProperty(propMap)
					if tp == itemSelectionEntry.PropertyTypeID {
						updatedProps = append(updatedProps, pb.buildGallerySelectionProperty(propMap, selectionMode))
					} else {
						updatedProps = append(updatedProps, propMap)
					}
				}
				result = append(result, bson.E{Key: "Properties", Value: updatedProps})
			} else {
				result = append(result, elem)
			}
		} else {
			result = append(result, elem)
		}
	}
	return result
}
