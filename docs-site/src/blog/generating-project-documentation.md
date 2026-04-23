# Generating Project Documentation with mxcli

*Draft — not yet published*

---

<!-- TODO: date -->

<!-- TODO: intro — turning a Mendix project into human-readable, shareable documentation -->

## mxcli report

```bash
mxcli report -p app.mpr --format markdown
mxcli report -p app.mpr --format html
```

<!-- TODO: what the report contains: scores, entity counts, microflow complexity, lint findings -->

## Demo

<!-- TODO: embed YouTube demo — generating a report and browsing it -->
<!--
<div style="position:relative;padding-bottom:56.25%;height:0;overflow:hidden;">
  <iframe style="position:absolute;top:0;left:0;width:100%;height:100%;"
    src="https://www.youtube.com/embed/VIDEO_ID"
    frameborder="0" allowfullscreen></iframe>
</div>
-->

## Building an mdBook Site from Your Project

<!-- TODO: workflow for generating an mdBook (or similar) from DESCRIBE output + mxcli report -->
<!-- TODO: example: auto-generated entity reference, microflow catalogue -->

## Keeping Docs in Sync

<!-- TODO: running doc generation in CI, committing the output or publishing to GitHub Pages -->

## Use Cases

<!-- TODO: onboarding new developers, audit trails, stakeholder handoffs -->
