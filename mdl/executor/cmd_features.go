// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/versions"
)

// execShowFeatures handles SHOW FEATURES, SHOW FEATURES FOR VERSION, and
// SHOW FEATURES ADDED SINCE commands.
func (e *Executor) execShowFeatures(s *ast.ShowFeaturesStmt) error {
	reg, err := versions.Load()
	if err != nil {
		return fmt.Errorf("failed to load version registry: %w", err)
	}

	// Determine the project version to use.
	var pv versions.SemVer

	switch {
	case s.AddedSince != "":
		// SHOW FEATURES ADDED SINCE x.y
		sinceV, err := versions.ParseSemVer(s.AddedSince)
		if err != nil {
			return fmt.Errorf("invalid version %q: %w", s.AddedSince, err)
		}
		return e.showFeaturesAddedSince(reg, sinceV)

	case s.ForVersion != "":
		// SHOW FEATURES FOR VERSION x.y — no project connection needed
		pv, err = versions.ParseSemVer(s.ForVersion)
		if err != nil {
			return fmt.Errorf("invalid version %q: %w", s.ForVersion, err)
		}

	default:
		// SHOW FEATURES [IN area] — requires project connection
		if e.reader == nil {
			return fmt.Errorf("not connected to a project\n  hint: use SHOW FEATURES FOR VERSION x.y without a project connection")
		}
		rpv := e.reader.ProjectVersion()
		pv = versions.SemVer{Major: rpv.MajorVersion, Minor: rpv.MinorVersion, Patch: rpv.PatchVersion}
	}

	if s.InArea != "" {
		return e.showFeaturesInArea(reg, pv, s.InArea)
	}
	return e.showFeaturesAll(reg, pv)
}

func (e *Executor) showFeaturesAll(reg *versions.Registry, pv versions.SemVer) error {
	features := reg.FeaturesForVersion(pv)
	if len(features) == 0 {
		fmt.Fprintf(e.output, "No features found for version %s\n", pv)
		return nil
	}

	fmt.Fprintf(e.output, "Features for Mendix %s:\n\n", pv)
	fmt.Fprintf(e.output, "| %-30s | %-9s | %-8s | %s |\n", "Feature", "Available", "Since", "Notes")
	fmt.Fprintf(e.output, "|%s|%s|%s|%s|\n",
		strings.Repeat("-", 32), strings.Repeat("-", 11), strings.Repeat("-", 10), strings.Repeat("-", 40))

	available, unavailable := 0, 0
	for _, f := range features {
		avail := "Yes"
		if !f.Available {
			avail = "No"
			unavailable++
		} else {
			available++
		}
		notes := f.Notes
		if !f.Available && f.Workaround != nil {
			notes = f.Workaround.Description
		}
		if len(notes) > 38 {
			notes = notes[:35] + "..."
		}
		fmt.Fprintf(e.output, "| %-30s | %-9s | %-8s | %-38s |\n",
			f.DisplayName(), avail, f.MinVersion, notes)
	}

	fmt.Fprintf(e.output, "\n(%d available, %d not available in %s)\n", available, unavailable, pv)
	return nil
}

func (e *Executor) showFeaturesInArea(reg *versions.Registry, pv versions.SemVer, area string) error {
	features := reg.FeaturesInArea(area, pv)
	if len(features) == 0 {
		// Check if the area exists at all.
		areas := reg.Areas()
		fmt.Fprintf(e.output, "No features found in area %q for version %s\n", area, pv)
		fmt.Fprintf(e.output, "Available areas: %s\n", strings.Join(areas, ", "))
		return nil
	}

	fmt.Fprintf(e.output, "Features in %s for Mendix %s:\n\n", area, pv)
	fmt.Fprintf(e.output, "| %-30s | %-9s | %-8s | %s |\n", "Feature", "Available", "Since", "Notes")
	fmt.Fprintf(e.output, "|%s|%s|%s|%s|\n",
		strings.Repeat("-", 32), strings.Repeat("-", 11), strings.Repeat("-", 10), strings.Repeat("-", 40))

	for _, f := range features {
		avail := "Yes"
		if !f.Available {
			avail = "No"
		}
		notes := f.Notes
		if !f.Available && f.Workaround != nil {
			notes = f.Workaround.Description
		}
		if len(notes) > 38 {
			notes = notes[:35] + "..."
		}
		fmt.Fprintf(e.output, "| %-30s | %-9s | %-8s | %-38s |\n",
			f.DisplayName(), avail, f.MinVersion, notes)
	}

	return nil
}

func (e *Executor) showFeaturesAddedSince(reg *versions.Registry, sinceV versions.SemVer) error {
	added := reg.FeaturesAddedSince(sinceV)
	if len(added) == 0 {
		fmt.Fprintf(e.output, "No new features found since %s\n", sinceV)
		return nil
	}

	fmt.Fprintf(e.output, "Features added since Mendix %s:\n\n", sinceV)
	fmt.Fprintf(e.output, "| %-30s | %-12s | %-10s | %s |\n", "Feature", "Area", "Since", "Notes")
	fmt.Fprintf(e.output, "|%s|%s|%s|%s|\n",
		strings.Repeat("-", 32), strings.Repeat("-", 14), strings.Repeat("-", 12), strings.Repeat("-", 40))

	for _, f := range added {
		notes := f.Notes
		if f.MDL != "" && notes == "" {
			notes = f.MDL
		}
		if len(notes) > 38 {
			notes = notes[:35] + "..."
		}
		fmt.Fprintf(e.output, "| %-30s | %-12s | %-10s | %-38s |\n",
			f.DisplayName(), f.Area, f.MinVersion, notes)
	}

	fmt.Fprintf(e.output, "\n(%d features added since %s)\n", len(added), sinceV)
	return nil
}
