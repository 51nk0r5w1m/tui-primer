/**
 * BubbleTea (Go) export pipeline for TUI Studio.
 *
 * Generates a .zip archive containing:
 *   myapp/
 *     go.mod            – module with replace directive → ./bubblestudio
 *     main.go           – minimal wiring, calls bubblestudio.Load
 *     handlers.go       – empty callback stubs for the user to fill in
 *     design.tui        – original TUI Studio design JSON
 *     bubblestudio/     – embedded runtime library
 *       go.mod
 *       bubblestudio.go
 *       components.go
 */

import type { ComponentNode } from '../../types';

// ── Embedded Go library sources (loaded at build time via Vite ?raw) ──────────

import libGoMod from '../../../bubblestudio/go.mod?raw';
import libBubblestudioGo from '../../../bubblestudio/bubblestudio.go?raw';
import libComponentsGo from '../../../bubblestudio/components.go?raw';

// ── CRC-32 ────────────────────────────────────────────────────────────────────

const CRC32_TABLE = (() => {
  const table = new Uint32Array(256);
  for (let i = 0; i < 256; i++) {
    let c = i;
    for (let j = 0; j < 8; j++) {
      c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
    }
    table[i] = c;
  }
  return table;
})();

function crc32(data: Uint8Array): number {
  let crc = 0xffffffff;
  for (let i = 0; i < data.length; i++) {
    crc = CRC32_TABLE[(crc ^ data[i]) & 0xff] ^ (crc >>> 8);
  }
  return (crc ^ 0xffffffff) >>> 0;
}

// ── ZIP encoder (STORE method, no compression) ────────────────────────────────

interface ZipFile {
  name: string;
  data: Uint8Array;
}

function encodeUtf8(s: string): Uint8Array {
  return new TextEncoder().encode(s);
}

function toBytes(data: string | Uint8Array): Uint8Array {
  return typeof data === 'string' ? encodeUtf8(data) : data;
}

function writeUint16LE(buf: Uint8Array, offset: number, value: number): void {
  buf[offset] = value & 0xff;
  buf[offset + 1] = (value >> 8) & 0xff;
}

function writeUint32LE(buf: Uint8Array, offset: number, value: number): void {
  buf[offset] = value & 0xff;
  buf[offset + 1] = (value >> 8) & 0xff;
  buf[offset + 2] = (value >> 16) & 0xff;
  buf[offset + 3] = (value >> 24) & 0xff;
}

/**
 * Builds a store-only (no compression) ZIP archive.
 * Accepts files as { name, data } pairs where data is string or Uint8Array.
 */
export function makeZip(files: Array<{ name: string; data: string | Uint8Array }>): Uint8Array {
  const prepared: ZipFile[] = files.map((f) => ({ name: f.name, data: toBytes(f.data) }));

  // First pass: calculate total size.
  const LOCAL_HEADER = 30;
  const CENTRAL_HEADER = 46;
  let localSize = 0;
  for (const f of prepared) {
    localSize += LOCAL_HEADER + f.name.length + f.data.length;
  }
  const centralSize = prepared.reduce((sum, f) => sum + CENTRAL_HEADER + f.name.length, 0);
  const totalSize = localSize + centralSize + 22; // 22 = EOCD

  const buf = new Uint8Array(totalSize);
  let pos = 0;

  // Local file records + remember offsets for central directory.
  const offsets: number[] = [];

  for (const f of prepared) {
    offsets.push(pos);
    const nameBytes = encodeUtf8(f.name);
    const checksum = crc32(f.data);
    const size = f.data.length;

    // Local file header signature
    buf[pos] = 0x50; buf[pos + 1] = 0x4b; buf[pos + 2] = 0x03; buf[pos + 3] = 0x04;
    writeUint16LE(buf, pos + 4, 20);   // version needed: 2.0
    writeUint16LE(buf, pos + 6, 0);    // general purpose bit flag
    writeUint16LE(buf, pos + 8, 0);    // compression method: STORE
    writeUint16LE(buf, pos + 10, 0);   // last mod time
    writeUint16LE(buf, pos + 12, 0);   // last mod date
    writeUint32LE(buf, pos + 14, checksum);
    writeUint32LE(buf, pos + 18, size); // compressed size
    writeUint32LE(buf, pos + 22, size); // uncompressed size
    writeUint16LE(buf, pos + 26, nameBytes.length);
    writeUint16LE(buf, pos + 28, 0);   // extra field length
    pos += 30;
    buf.set(nameBytes, pos); pos += nameBytes.length;
    buf.set(f.data, pos);   pos += size;
  }

  const centralStart = pos;

  // Central directory.
  for (let i = 0; i < prepared.length; i++) {
    const f = prepared[i];
    const nameBytes = encodeUtf8(f.name);
    const checksum = crc32(f.data);
    const size = f.data.length;
    const offset = offsets[i];

    buf[pos] = 0x50; buf[pos + 1] = 0x4b; buf[pos + 2] = 0x01; buf[pos + 3] = 0x02;
    writeUint16LE(buf, pos + 4, 20);   // version made by
    writeUint16LE(buf, pos + 6, 20);   // version needed
    writeUint16LE(buf, pos + 8, 0);    // general purpose bit flag
    writeUint16LE(buf, pos + 10, 0);   // compression method: STORE
    writeUint16LE(buf, pos + 12, 0);   // last mod time
    writeUint16LE(buf, pos + 14, 0);   // last mod date
    writeUint32LE(buf, pos + 16, checksum);
    writeUint32LE(buf, pos + 20, size); // compressed size
    writeUint32LE(buf, pos + 24, size); // uncompressed size
    writeUint16LE(buf, pos + 28, nameBytes.length);
    writeUint16LE(buf, pos + 30, 0);   // extra field length
    writeUint16LE(buf, pos + 32, 0);   // file comment length
    writeUint16LE(buf, pos + 34, 0);   // disk number start
    writeUint16LE(buf, pos + 36, 0);   // internal file attributes
    writeUint32LE(buf, pos + 38, 0);   // external file attributes
    writeUint32LE(buf, pos + 42, offset);
    pos += 46;
    buf.set(nameBytes, pos); pos += nameBytes.length;
  }

  // End of central directory record.
  const centralLen = pos - centralStart;
  buf[pos] = 0x50; buf[pos + 1] = 0x4b; buf[pos + 2] = 0x05; buf[pos + 3] = 0x06;
  writeUint16LE(buf, pos + 4, 0);                   // disk number
  writeUint16LE(buf, pos + 6, 0);                   // disk with central dir
  writeUint16LE(buf, pos + 8, prepared.length);     // entries on disk
  writeUint16LE(buf, pos + 10, prepared.length);    // total entries
  writeUint32LE(buf, pos + 12, centralLen);          // central dir size
  writeUint32LE(buf, pos + 16, centralStart);        // central dir offset
  writeUint16LE(buf, pos + 20, 0);                   // comment length

  return buf;
}

// ── Identifier helpers ────────────────────────────────────────────────────────

/**
 * Derives a stable, exported Go identifier from a component name and type.
 * Falls back to <Type><index> if the name is empty, numeric, or duplicates
 * a reserved word.
 */
export function toGoIdent(name: string, type: string, index: number): string {
  const GO_RESERVED = new Set([
    'break', 'case', 'chan', 'const', 'continue', 'default', 'defer', 'else',
    'fallthrough', 'for', 'func', 'go', 'goto', 'if', 'import', 'interface',
    'map', 'package', 'range', 'return', 'select', 'struct', 'switch', 'type', 'var',
  ]);

  if (!name || /^\d+$/.test(name)) {
    return `${type}${index}`;
  }

  // CamelCase: replace non-alphanum word boundaries with uppercase.
  const camel = name
    .replace(/[^a-zA-Z0-9]+(.)/g, (_, c: string) => c.toUpperCase())
    .replace(/^[^a-zA-Z_]/, '_')
    // Ensure first char is uppercase (exported identifier).
    .replace(/^(.)/, (c: string) => c.toUpperCase());

  if (GO_RESERVED.has(camel.toLowerCase())) {
    return `${camel}${type}`;
  }

  return camel || `${type}${index}`;
}

// ── Tree walking helpers ──────────────────────────────────────────────────────

const INTERACTIVE_TYPES = new Set([
  'TextInput', 'Button', 'Checkbox', 'Radio', 'Toggle', 'Select', 'Tabs',
]);

interface InteractiveComp {
  node: ComponentNode;
  ident: string;
}

function collectInteractive(
  root: ComponentNode | null,
  out: InteractiveComp[] = []
): InteractiveComp[] {
  if (!root || root.hidden) return out;
  if (INTERACTIVE_TYPES.has(root.type)) {
    const ident = toGoIdent(root.name, root.type, out.length);
    out.push({ node: root, ident });
  }
  for (const child of root.children) {
    collectInteractive(child, out);
  }
  return out;
}

// ── Generated file content ────────────────────────────────────────────────────

/** Returns the content of main.go — the minimal wiring file. */
export function exportToBubbleTeaMain(_root: ComponentNode | null): string {
  return `package main

import (
\t"fmt"
\t"os"

\ttea "github.com/charmbracelet/bubbletea"
\t"github.com/tuistudio/bubblestudio"
)

func main() {
\tm, err := bubblestudio.Load("design.tui", registerHandlers())
\tif err != nil {
\t\tfmt.Fprintln(os.Stderr, "Error loading design:", err)
\t\tos.Exit(1)
\t}
\tp := tea.NewProgram(m, tea.WithAltScreen())
\tif _, err := p.Run(); err != nil {
\t\tfmt.Fprintln(os.Stderr, "Error running program:", err)
\t\tos.Exit(1)
\t}
}
`;
}

/** Returns the content of handlers.go — empty stubs the user fills in. */
export function exportBubbleTeaHandlers(root: ComponentNode | null): string {
  const interactive = collectInteractive(root);

  const lines: string[] = [
    'package main',
    '',
    'import "github.com/tuistudio/bubblestudio"',
    '',
    '// registerHandlers wires your application logic to UI events.',
    '// Fill in the functions below; leave them empty if unused.',
    'func registerHandlers() bubblestudio.Handlers {',
    '\treturn bubblestudio.Handlers{',
  ];

  // OnClick (Button)
  const buttons = interactive.filter((c) => c.node.type === 'Button');
  if (buttons.length > 0) {
    lines.push('\t\tOnClick: map[string]func(){');
    for (const { node, ident } of buttons) {
      lines.push(`\t\t\t// ${node.name || node.type} button`);
      lines.push(`\t\t\t"${node.name}": on${ident}Click,`);
    }
    lines.push('\t\t},');
  } else {
    lines.push('\t\tOnClick: map[string]func(){},');
  }

  // OnChange / OnSubmit (TextInput)
  const inputs = interactive.filter((c) => c.node.type === 'TextInput');
  if (inputs.length > 0) {
    lines.push('\t\tOnChange: map[string]func(string){');
    for (const { node, ident } of inputs) {
      lines.push(`\t\t\t"${node.name}": on${ident}Change,`);
    }
    lines.push('\t\t},');
    lines.push('\t\tOnSubmit: map[string]func(string){');
    for (const { node, ident } of inputs) {
      lines.push(`\t\t\t"${node.name}": on${ident}Submit,`);
    }
    lines.push('\t\t},');
  } else {
    lines.push('\t\tOnChange: map[string]func(string){},');
    lines.push('\t\tOnSubmit: map[string]func(string){},');
  }

  // OnToggle (Checkbox, Radio, Toggle)
  const toggles = interactive.filter((c) =>
    ['Checkbox', 'Radio', 'Toggle'].includes(c.node.type)
  );
  if (toggles.length > 0) {
    lines.push('\t\tOnToggle: map[string]func(bool){');
    for (const { node, ident } of toggles) {
      lines.push(`\t\t\t"${node.name}": on${ident}Toggle,`);
    }
    lines.push('\t\t},');
  } else {
    lines.push('\t\tOnToggle: map[string]func(bool){},');
  }

  // OnSelect (Select)
  const selects = interactive.filter((c) => c.node.type === 'Select');
  if (selects.length > 0) {
    lines.push('\t\tOnSelect: map[string]func(string){');
    for (const { node, ident } of selects) {
      lines.push(`\t\t\t"${node.name}": on${ident}Select,`);
    }
    lines.push('\t\t},');
  } else {
    lines.push('\t\tOnSelect: map[string]func(string){},');
  }

  // OnTab (Tabs)
  const tabs = interactive.filter((c) => c.node.type === 'Tabs');
  if (tabs.length > 0) {
    lines.push('\t\tOnTab: map[string]func(string){');
    for (const { node, ident } of tabs) {
      lines.push(`\t\t\t"${node.name}": on${ident}Tab,`);
    }
    lines.push('\t\t},');
  } else {
    lines.push('\t\tOnTab: map[string]func(string){},');
  }

  lines.push('\t}');
  lines.push('}');
  lines.push('');

  // Emit stub functions.
  for (const { node, ident } of buttons) {
    lines.push(`// on${ident}Click is called when the "${node.name}" button is activated.`);
    lines.push(`func on${ident}Click() {`);
    lines.push(`\t// TODO: handle ${node.name} click`);
    lines.push('}');
    lines.push('');
  }

  for (const { node, ident } of inputs) {
    lines.push(`// on${ident}Change is called each time the "${node.name}" input changes.`);
    lines.push(`func on${ident}Change(value string) {`);
    lines.push(`\t_ = value // TODO: handle ${node.name} change`);
    lines.push('}');
    lines.push('');
    lines.push(`// on${ident}Submit is called when the "${node.name}" input is submitted.`);
    lines.push(`func on${ident}Submit(value string) {`);
    lines.push(`\t_ = value // TODO: handle ${node.name} submit`);
    lines.push('}');
    lines.push('');
  }

  for (const { node, ident } of toggles) {
    lines.push(`// on${ident}Toggle is called when "${node.name}" changes.`);
    lines.push(`func on${ident}Toggle(checked bool) {`);
    lines.push(`\t_ = checked // TODO: handle ${node.name} toggle`);
    lines.push('}');
    lines.push('');
  }

  for (const { node, ident } of selects) {
    lines.push(`// on${ident}Select is called when an item is chosen from "${node.name}".`);
    lines.push(`func on${ident}Select(value string) {`);
    lines.push(`\t_ = value // TODO: handle ${node.name} select`);
    lines.push('}');
    lines.push('');
  }

  for (const { node, ident } of tabs) {
    lines.push(`// on${ident}Tab is called when the active tab in "${node.name}" changes.`);
    lines.push(`func on${ident}Tab(label string) {`);
    lines.push(`\t_ = label // TODO: handle ${node.name} tab change`);
    lines.push('}');
    lines.push('');
  }

  return lines.join('\n');
}

/** Returns the content of the app-level go.mod. */
function appGoMod(): string {
  return `module myapp

go 1.22

require github.com/tuistudio/bubblestudio v0.0.0

replace github.com/tuistudio/bubblestudio => ./bubblestudio
`;
}

// ── Public API ────────────────────────────────────────────────────────────────

/**
 * Builds and returns a zip archive as a Uint8Array.
 *
 * @param root      The root ComponentNode from the design (may be null).
 * @param tuiJson   The raw .tui JSON string to embed in the archive.
 */
export function exportToBubbleTeaZip(
  root: ComponentNode | null,
  tuiJson: string
): Uint8Array {
  const files: Array<{ name: string; data: string }> = [
    { name: 'myapp/go.mod',                         data: appGoMod() },
    { name: 'myapp/main.go',                        data: exportToBubbleTeaMain(root) },
    { name: 'myapp/handlers.go',                    data: exportBubbleTeaHandlers(root) },
    { name: 'myapp/design.tui',                     data: tuiJson },
    { name: 'myapp/bubblestudio/go.mod',             data: libGoMod },
    { name: 'myapp/bubblestudio/bubblestudio.go',    data: libBubblestudioGo },
    { name: 'myapp/bubblestudio/components.go',      data: libComponentsGo },
  ];

  return makeZip(files);
}
