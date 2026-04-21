# SVG Sketch Style Guide — PoC Diagrams

**When to use:** Apply this style to ALL hand-built SVG diagrams (ELK system overview, any future SVG rendered in webviews). This communicates "proof of concept / draft" and prevents stakeholders from perceiving diagrams as final/productized.

**Where SVGs are generated:** `vscode-mdl/src/previewProvider.ts` in the `renderSvg()` function and `getElkWebviewContent()`. Any new diagram type that builds SVG strings must follow this guide.

## Core Principles

- Every line wobbles slightly — nothing is perfectly straight
- Fills use horizontal marker strokes, not solid `fill` colors
- Transparent SVG background (inherits from VS Code theme)
- Hand-drawn font (Architects Daughter from Google Fonts)
- Muted, slightly desaturated palette — no pure black (use `#2c3e50` for ink, `#5a5a5a` for secondary)
- Always include a "PoC draft" or "sketch — subject to change" label

## Google Font Import

Include this in the HTML `<head>` of any webview that renders sketch SVGs:

```html
<link href="https://fonts.googleapis.com/css2?family=Architects+Daughter&display=swap" rel="stylesheet">
```

## SVG Filter Definitions

Include these three filters in every SVG `<defs>`:

```xml
<!-- Pencil wobble: displaces all strokes slightly -->
<filter id="pencil">
  <feTurbulence type="turbulence" baseFrequency="0.03" numOctaves="4" result="noise" />
  <feDisplacementMap in="SourceGraphic" in2="noise" scale="1.5" />
</filter>

<!-- Marker texture: directional noise that mimics marker grain -->
<filter id="marker-texture">
  <feTurbulence type="fractalNoise" baseFrequency="0.04 0.15" numOctaves="3" result="noise" />
  <feDisplacementMap in="SourceGraphic" in2="noise" scale="2" />
  <feGaussianBlur stdDeviation="0.3" />
</filter>
```

## Seeded PRNG

Use deterministic randomness so diagrams look the same on re-render but each shape has unique wobble:

```javascript
function makeRng(seed) {
  let s = seed || 1;
  return () => {
    s = (s * 16807) % 2147483647;
    return (s - 1) / 2147483646;
  };
}
```

Pass unique seeds per shape element (e.g., hash of node ID) so each box/line has its own consistent wobble pattern.

## Hand-Drawn Lines (roughLine)

Never use plain SVG `<rect>`, `<line>`, or `<circle>`. Instead, generate paths by interpolating between start/end points in small segments (~18px each), adding random jitter (perpendicular to the line direction) at each segment point. Skip jitter on the final point so endpoints land accurately.

```javascript
function roughLine(x1, y1, x2, y2, rng, jitter) {
  jitter = jitter || 1.5;
  const dx = x2 - x1, dy = y2 - y1;
  const len = Math.sqrt(dx * dx + dy * dy);
  const segments = Math.max(Math.ceil(len / 18), 2);
  // perpendicular unit vector
  const px = -dy / len, py = dx / len;
  let d = 'M ' + x1 + ' ' + y1;
  for (let i = 1; i <= segments; i++) {
    const t = i / segments;
    let x = x1 + dx * t;
    let y = y1 + dy * t;
    if (i < segments) { // no jitter on last point
      const j = (rng() - 0.5) * 2 * jitter;
      x += px * j;
      y += py * j;
    }
    d += ' L ' + x.toFixed(1) + ' ' + y.toFixed(1);
  }
  return d;
}
```

Apply `filter="url(#pencil)"` and `stroke-linecap="round"` to all shape outlines.

### Rough Rounded Rectangle

```javascript
function roughRoundedRect(x, y, w, h, r, rng) {
  const j = 1.5;
  // top edge (after top-left corner to before top-right corner)
  let d = 'M ' + (x + r) + ' ' + y;
  d += roughLineSegments(x + r, y, x + w - r, y, rng, j);
  // top-right corner
  d += ' Q ' + (x + w) + ' ' + y + ' ' + (x + w) + ' ' + (y + r);
  // right edge
  d += roughLineSegments(x + w, y + r, x + w, y + h - r, rng, j);
  // bottom-right corner
  d += ' Q ' + (x + w) + ' ' + (y + h) + ' ' + (x + w - r) + ' ' + (y + h);
  // bottom edge
  d += roughLineSegments(x + w - r, y + h, x + r, y + h, rng, j);
  // bottom-left corner
  d += ' Q ' + x + ' ' + (y + h) + ' ' + x + ' ' + (y + h - r);
  // left edge
  d += roughLineSegments(x, y + h - r, x, y + r, rng, j);
  // top-left corner
  d += ' Q ' + x + ' ' + y + ' ' + (x + r) + ' ' + y;
  d += ' Z';
  return d;
}
```

## Marker Fill Technique

Instead of using `fill="color"` on shapes, draw semi-transparent horizontal strokes across the shape interior:

1. Loop from top to bottom of the bounding box in ~4-5px spacing
2. Each stroke is a quadratic bezier path with slight vertical jitter
3. Inset strokes 3px from edges so they don't bleed past the border
4. Draw **two passes** at slight offset for overlap effect:
   - Pass 1: `strokeWidth="4"`, `opacity="0.6-0.7"`
   - Pass 2: `strokeWidth="3.5"`, `opacity="0.3-0.5"`, offset by ~1px
5. Apply `filter="url(#marker-texture)"` and `stroke-linecap="round"` to fill strokes
6. Use the **light** variant of the color for fills, **base** variant for borders

```javascript
function markerFill(x, y, w, h, lightColor, rng) {
  const inset = 3;
  const spacing = 4.5;
  let paths = '';
  for (let pass = 0; pass < 2; pass++) {
    const sw = pass === 0 ? 4 : 3.5;
    const op = pass === 0 ? 0.65 : 0.4;
    const offsetY = pass * 1;
    for (let ly = y + inset + offsetY; ly < y + h - inset; ly += spacing) {
      const jx1 = (rng() - 0.5) * 2;
      const jx2 = (rng() - 0.5) * 2;
      const jy = (rng() - 0.5) * 4;
      const mx = x + w / 2 + jx1;
      const my = ly + jy;
      paths += '<path d="M ' + (x + inset) + ' ' + ly +
        ' Q ' + mx.toFixed(1) + ' ' + my.toFixed(1) + ' ' + (x + w - inset) + ' ' + (ly + jx2).toFixed(1) +
        '" fill="none" stroke="' + lightColor + '" stroke-width="' + sw +
        '" opacity="' + op + '" stroke-linecap="round" filter="url(#marker-texture)"/>';
    }
  }
  return paths;
}
```

For circles/diamonds, clip the horizontal stroke lengths to fit within the shape boundary.

## Color Palette

Each color has a `base` (border/header) and `light` (marker fill) variant:

| Role             | Base      | Light     |
|------------------|-----------|-----------|
| Blue (primary)   | `#4a90d9` | `#d0e1f9` |
| Green (success)  | `#5ba55b` | `#d4edda` |
| Orange (warning) | `#e89b3e` | `#fce4c0` |
| Purple (accent)  | `#8e6cbf` | `#e2d5f0` |
| Pink (error)     | `#d4618c` | `#f5d0de` |

| Element          | Color     |
|------------------|-----------|
| Ink (primary text)| `#2c3e50`|
| Pencil (secondary)| `#5a5a5a`|
| Connector lines  | `#6b7b8d` |
| True/yes path    | `#3a8a3a` |
| False/no path    | `#c0392b` |

### VS Code Theme Mapping

In webview CSS, map sketch colors to VS Code theme variables with fallbacks:

```javascript
const inkColor = 'var(--vscode-editor-foreground, #2c3e50)';
// No SVG background — inherits from VS Code theme
const connectorColor = '#6b7b8d';
```

## Typography

- **Font**: `'Architects Daughter', cursive` (Google Fonts)
- **Weight**: 400 only (single weight font)
- **Entity/box titles**: 12-14px
- **Subtitles/types**: 10-11px, secondary color (`#5a5a5a`)
- **Diagram title**: 20px
- **Annotations**: 10px, opacity 0.7
- **All SVG text**: `font-family="'Architects Daughter', cursive"`

## Shape Reference

| Shape | Use | Technique |
|-------|-----|-----------|
| Rounded rectangle | Activities, entities, modules | `roughRoundedRect()` + marker fill |
| Rectangle | Entity attribute area | `roughLine` the 4 sides |
| Diamond | Decisions, splits, merges | `roughLine` the 4 diagonal sides |
| Circle | Start/end events | Plot points around circumference with radius jitter |
| Double circle | End event | Two concentric rough circles |
| Dashed line | Optional associations, annotations | `stroke-dasharray="5 3"` on rough path |

## Arrows

- Path: use `roughLine` between waypoints (ELK bend points)
- Arrowhead: two short lines from tip at +/-0.35 radians, with slight jitter
- Connector color: `#6b7b8d`
- Labels: placed at midpoint, 10-11px, italic for association names
- Dashed arrows for annotations: `stroke-dasharray="5 3"`, lighter color (`#999`)

```javascript
function roughArrowhead(tipX, tipY, angle, rng) {
  const len = 10;
  const spread = 0.35;
  const j = (rng() - 0.5) * 1.5;
  const x1 = tipX - len * Math.cos(angle - spread) + j;
  const y1 = tipY - len * Math.sin(angle - spread) + j;
  const x2 = tipX - len * Math.cos(angle + spread) - j;
  const y2 = tipY - len * Math.sin(angle + spread) - j;
  return '<path d="M ' + x1.toFixed(1) + ' ' + y1.toFixed(1) +
    ' L ' + tipX + ' ' + tipY +
    ' L ' + x2.toFixed(1) + ' ' + y2.toFixed(1) +
    '" fill="none" stroke="#6b7b8d" stroke-width="1.5" stroke-linecap="round" filter="url(#pencil)"/>';
}
```

## Footer / PoC Indicators

Always include at least one of these in the SVG:

```javascript
// Badge near title
svg += '<rect x="' + (titleX + titleWidth + 8) + '" y="' + (titleY - 12) +
  '" width="60" height="18" rx="9" fill="none" stroke="#5a5a5a" stroke-width="1.5" filter="url(#pencil)"/>';
svg += '<text x="' + (titleX + titleWidth + 38) + '" y="' + (titleY + 1) +
  '" font-size="9" fill="#5a5a5a" text-anchor="middle" font-family="\'Architects Daughter\', cursive">PoC draft</text>';

// footer text
svg += '<text x="10" y="' + (svgHeight - 8) +
  '" font-size="10" fill="#5a5a5a" opacity="0.4" font-family="\'Architects Daughter\', cursive">sketch — subject to change</text>';
```

## Annotations

- Dashed rectangle border (`stroke-dasharray="4 3"`, `opacity="0.6"`)
- No fill
- Small text inside, 10px, secondary color

## Complete SVG Structure Template

```javascript
function buildSketchSvg(svgWidth, svgHeight, renderContent) {
  let svg = '<svg xmlns="http://www.w3.org/2000/svg" width="' + svgWidth + '" height="' + svgHeight + '">';

  // 1. filter definitions
  svg += '<defs>';
  svg += '<filter id="pencil"><feTurbulence type="turbulence" baseFrequency="0.03" numOctaves="4" result="noise"/><feDisplacementMap in="SourceGraphic" in2="noise" scale="1.5"/></filter>';
  svg += '<filter id="marker-texture"><feTurbulence type="fractalNoise" baseFrequency="0.04 0.15" numOctaves="3" result="noise"/><feDisplacementMap in="SourceGraphic" in2="noise" scale="2"/><feGaussianBlur stdDeviation="0.3"/></filter>';
  svg += '<filter id="grain"><feTurbulence type="fractalNoise" baseFrequency="1.5" numOctaves="1" result="grain"/><feColorMatrix in="grain" type="saturate" values="0" result="bw"/><feBlend in="SourceGraphic" in2="bw" mode="multiply" result="blended"/><feComponentTransfer in="blended"><feFuncA type="linear" slope="0.3"/></feComponentTransfer></filter>';
  svg += '</defs>';

  // 2. Paper background
  // No background rect — transparent SVG inherits from webview theme
  svg += '<rect width="' + svgWidth + '" height="' + svgHeight + '" filter="url(#grain)" opacity="0.4"/>';

  // 3. Diagram content (nodes, edges, labels)
  svg += renderContent();

  // 4. footer
  svg += '<text x="10" y="' + (svgHeight - 8) + '" font-size="10" fill="#5a5a5a" opacity="0.4" font-family="\'Architects Daughter\', cursive">sketch — subject to change</text>';

  svg += '</svg>';
  return svg;
}
```

## Implementation Checklist

When creating or modifying SVG rendering code:

- [ ] Include Google Font link in webview HTML `<head>`
- [ ] Add all three SVG filters (`pencil`, `marker-texture`, `grain`) in `<defs>`
- [ ] Replace all `<rect>` with `roughRoundedRect()` paths
- [ ] Replace all `<line>` with `roughLine()` paths
- [ ] Replace solid `fill` colors with `markerFill()` horizontal strokes
- [ ] Leave SVG background transparent (inherits from VS Code theme)
- [ ] Set all text to `font-family="'Architects Daughter', cursive"`
- [ ] Use ink color `#2c3e50` for primary text, `#5a5a5a` for secondary
- [ ] Use `#6b7b8d` for connector lines
- [ ] Add PoC badge or footer text
- [ ] Use seeded PRNG (`makeRng`) for deterministic wobble
- [ ] Apply `filter="url(#pencil)"` to shape outlines
- [ ] Apply `filter="url(#marker-texture)"` to marker fill strokes
- [ ] Use `stroke-linecap="round"` on all paths
