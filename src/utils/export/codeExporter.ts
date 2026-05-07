// Code generation for different TUI frameworks

import type { ComponentNode } from '../../types';
import { exportToBubbleTeaMain } from './bubbletea';

/**
 * Export design to framework-specific code
 */
export function exportToCode(root: ComponentNode | null, format: string): string {
  if (!root) return '';

  switch (format) {
    case 'opentui':
      return exportToOpenTUI(root);
    case 'ink':
      return exportToInk(root);
    case 'bubbletea':
      return exportToBubbleTea(root);
    case 'blessed':
      return exportToBlessed(root);
    case 'textual':
      return exportToTextual(root);
    default:
      return `// Unsupported export format: ${format}`;
  }
}

// ── Ink ───────────────────────────────────────────────────────────────────────

function exportToInk(root: ComponentNode): string {
  const extras = new Set<string>();
  collectInkImports(root, extras);

  let importLines = `import React from 'react';\nimport { render, Box, Text } from 'ink';`;
  if (extras.has('TextInput')) importLines += `\nimport TextInput from 'ink-text-input';`;
  if (extras.has('SelectInput')) importLines += `\nimport SelectInput from 'ink-select-input';`;
  if (extras.has('Spinner')) importLines += `\nimport Spinner from 'ink-spinner';`;

  const packageNote = buildPackageNote(extras);

  return `${importLines}

function App() {
  return (
${generateInkNode(root, 2)}  );
}

render(<App />);
${packageNote}`;
}

function buildPackageNote(extras: Set<string>): string {
  const pkgMap: Record<string, string> = {
    TextInput: 'ink-text-input',
    SelectInput: 'ink-select-input',
    Spinner: 'ink-spinner',
  };
  const needed = Array.from(extras)
    .map((k) => pkgMap[k])
    .filter(Boolean);
  if (!needed.length) return '';
  return `\n// Install extra packages:\n// npm install ${needed.join(' ')}\n`;
}

function collectInkImports(node: ComponentNode, imports: Set<string>): void {
  if (node.type === 'TextInput') imports.add('TextInput');
  if (node.type === 'Select') imports.add('SelectInput');
  if (node.type === 'Spinner') imports.add('Spinner');
  for (const child of node.children) collectInkImports(child, imports);
}

function generateInkNode(node: ComponentNode, indent: number): string {
  if (node.hidden) return '';
  const sp = '  '.repeat(indent);

  switch (node.type) {
    case 'Screen':
    case 'Box':
    case 'Grid':
    case 'Modal': {
      const props = inkBoxProps(node);
      const children = node.children.map((c) => generateInkNode(c, indent + 1)).join('');
      return children ? `${sp}<Box${props}>\n${children}${sp}</Box>\n` : `${sp}<Box${props} />\n`;
    }

    case 'Spacer':
      return `${sp}<Box flexGrow={1} />\n`;

    case 'Text': {
      const content = (node.props.content as string) || '';
      return `${sp}<Text${inkTextProps(node)}>${escJsx(content)}</Text>\n`;
    }

    case 'Button': {
      const label = (node.props.label as string) || 'Button';
      return `${sp}<Text${inkTextProps(node)} bold inverse>  ${escJsx(label)}  </Text>\n`;
    }

    case 'TextInput': {
      const placeholder = JSON.stringify((node.props.placeholder as string) || '');
      const value = JSON.stringify((node.props.value as string) || '');
      return `${sp}<TextInput value={${value}} placeholder={${placeholder}} onChange={() => {}} />\n`;
    }

    case 'Checkbox': {
      const label = (node.props.label as string) || '';
      const checked = !!node.props.checked;
      const icon = checked
        ? (node.props.checkedIcon as string) || '✓'
        : (node.props.uncheckedIcon as string) || '○';
      return `${sp}<Text${inkTextProps(node)}>{/* checked={${checked}} */} ${escJsx(icon)} ${escJsx(label)}</Text>\n`;
    }

    case 'Radio': {
      const label = (node.props.label as string) || '';
      const selected = !!node.props.checked;
      const icon = selected
        ? (node.props.selectedIcon as string) || '◉'
        : (node.props.unselectedIcon as string) || '○';
      return `${sp}<Text${inkTextProps(node)}>{/* selected={${selected}} */} ${escJsx(icon)} ${escJsx(label)}</Text>\n`;
    }

    case 'Toggle': {
      const label = (node.props.label as string) || '';
      const on = !!node.props.value;
      return `${sp}<Text${inkTextProps(node)}>{/* on={${on}} */} {${on} ? '[ON ]' : '[OFF]'} ${escJsx(label)}</Text>\n`;
    }

    case 'Select': {
      const options = (node.props.options as string[]) || ['Option 1', 'Option 2'];
      const items = options
        .map(
          (o: string) =>
            `{ label: ${JSON.stringify(o)}, value: ${JSON.stringify(o.toLowerCase().replace(/\s+/g, '_'))} }`
        )
        .join(', ');
      return `${sp}<SelectInput items={[${items}]} onSelect={() => {}} />\n`;
    }

    case 'Spinner':
      return `${sp}<Text${inkTextProps(node)}><Spinner type="dots" /></Text>\n`;

    case 'ProgressBar': {
      const value = (node.props.value as number) ?? 0;
      const max = (node.props.max as number) ?? 100;
      const width = (node.props.width as number) ?? 20;
      return (
        `${sp}<Text${inkTextProps(node)}>\n` +
        `${sp}  {'█'.repeat(Math.round(${value} / ${max} * ${width}))}` +
        `{'░'.repeat(${width} - Math.round(${value} / ${max} * ${width}))} ${value}%\n` +
        `${sp}</Text>\n`
      );
    }

    case 'List': {
      const items = (node.props.items as any[]) || [];
      const rows = items
        .map((item: any) => {
          const d = typeof item === 'string' ? { label: item, icon: '•' } : item;
          return `${sp}  <Text key={${JSON.stringify(d.label)}}>${escJsx(d.icon || '•')} ${escJsx(d.label)}</Text>`;
        })
        .join('\n');
      return `${sp}<Box${inkBoxProps(node)} flexDirection="column">\n${rows}\n${sp}</Box>\n`;
    }

    case 'Menu': {
      const items = (node.props.items as any[]) || [];
      const isRow = (node.layout as any).direction === 'row';
      const rows = items
        .map((item: any) => {
          const d = typeof item === 'string' ? { label: item, icon: '' } : item;
          const prefix = d.icon ? `${escJsx(d.icon)} ` : '';
          return `${sp}  <Text key={${JSON.stringify(d.label)}}>${prefix}${escJsx(d.label)}</Text>`;
        })
        .join('\n');
      return `${sp}<Box${inkBoxProps(node)} flexDirection="${isRow ? 'row' : 'column'}" gap={1}>\n${rows}\n${sp}</Box>\n`;
    }

    case 'Tabs': {
      const tabs = (node.props.tabs as any[]) || [];
      const rows = tabs
        .map((tab: any) => {
          const label = typeof tab === 'string' ? tab : tab.label || 'Tab';
          return `${sp}  <Text key={${JSON.stringify(label)}} underline> ${escJsx(label)} </Text>`;
        })
        .join('\n');
      return `${sp}<Box${inkBoxProps(node)} flexDirection="row">\n${rows}\n${sp}</Box>\n`;
    }

    case 'Table': {
      const columns = (node.props.columns as string[]) || ['Column 1', 'Column 2'];
      const rows = (node.props.rows as string[][]) || [];
      const colW = 14;
      const header = columns.map((c: string) => c.slice(0, colW).padEnd(colW)).join(' │ ');
      const divider = columns.map(() => '─'.repeat(colW)).join('─┼─');
      const dataRows = rows.map((row: string[]) =>
        columns
          .map((_: string, ci: number) => (row[ci] || '').slice(0, colW).padEnd(colW))
          .join(' │ ')
      );
      const lines = [header, divider, ...dataRows]
        .map((l, i) => `${sp}  <Text key={${i}}>{${JSON.stringify(l)}}</Text>`)
        .join('\n');
      return `${sp}<Box${inkBoxProps(node)} flexDirection="column">\n${lines}\n${sp}</Box>\n`;
    }

    case 'Tree': {
      const items = (node.props.items as any[]) || [];
      const flatLines: string[] = [];
      const walk = (item: any, depth: number) => {
        const d = typeof item === 'string' ? { label: item, children: [] } : item;
        const pad = '  '.repeat(depth) + (depth > 0 ? '├─ ' : '');
        flatLines.push(
          `${sp}  <Text key={${JSON.stringify(pad + d.label)}}>{${JSON.stringify(pad + d.label)}}</Text>`
        );
        (d.children || []).forEach((child: any) => walk(child, depth + 1));
      };
      items.forEach((item: any) => walk(item, 0));
      return `${sp}<Box${inkBoxProps(node)} flexDirection="column">\n${flatLines.join('\n')}\n${sp}</Box>\n`;
    }

    case 'Breadcrumb': {
      const items = (node.props.items as any[]) || [];
      const separator = (node.props.separator as string) || '/';
      const text = items
        .map((i: any) => (typeof i === 'string' ? i : i.label || ''))
        .join(` ${separator} `);
      return `${sp}<Text${inkTextProps(node)}>{${JSON.stringify(text)}}</Text>\n`;
    }

    default:
      return `${sp}{/* ${node.type}: ${escJsx(node.name)} */}\n`;
  }
}

/** Box-level props: flexbox layout + border from node.layout + node.style */
function inkBoxProps(node: ComponentNode): string {
  const props: string[] = [];
  const layout = node.layout as any;

  if (layout.direction === 'row') props.push('flexDirection="row"');
  if (layout.gap > 0) props.push(`gap={${layout.gap}}`);
  if (layout.padding > 0) props.push(`padding={${layout.padding}}`);

  const jMap: Record<string, string> = {
    center: 'center',
    end: 'flex-end',
    'space-between': 'space-between',
    between: 'space-between',
    'space-around': 'space-around',
    around: 'space-around',
    'space-evenly': 'space-evenly',
    evenly: 'space-evenly',
  };
  if (layout.justify && jMap[layout.justify])
    props.push(`justifyContent="${jMap[layout.justify]}"`);

  const aMap: Record<string, string> = { center: 'center', end: 'flex-end' };
  if (layout.align && aMap[layout.align]) props.push(`alignItems="${aMap[layout.align]}"`);

  if (typeof layout.width === 'number' && layout.width > 0) props.push(`width={${layout.width}}`);
  else if (layout.width === 'fill' || layout.width === 'fill_container') props.push('flexGrow={1}');
  if (typeof layout.height === 'number' && layout.height > 0)
    props.push(`height={${layout.height}}`);

  if (node.style.border) {
    const bsMap: Record<string, string> = {
      single: 'single',
      double: 'double',
      round: 'round',
      bold: 'bold',
      classic: 'classic',
    };
    const bs = bsMap[(node.style.borderStyle as string) || 'single'] || 'single';
    props.push(`borderStyle="${bs}"`);
    if (node.style.color) props.push(`borderColor="${node.style.color}"`);
  }

  return props.length ? ' ' + props.join(' ') : '';
}

/** Text-level props: color, bold, italic, underline from node.style */
function inkTextProps(node: ComponentNode): string {
  const props: string[] = [];
  if (node.style.color) props.push(`color="${node.style.color}"`);
  if (node.style.backgroundColor) props.push(`backgroundColor="${node.style.backgroundColor}"`);
  if (node.style.bold) props.push('bold');
  if (node.style.italic) props.push('italic');
  if (node.style.underline) props.push('underline');
  if ((node.style as any).strikethrough) props.push('strikethrough');
  return props.length ? ' ' + props.join(' ') : '';
}

/** Escape characters that are special in JSX text content */
function escJsx(s: string): string {
  return s.replace(
    /[{}<>&]/g,
    (c) => ({ '{': '&#123;', '}': '&#125;', '<': '&lt;', '>': '&gt;', '&': '&amp;' })[c] || c
  );
}

// ── OpenTUI ───────────────────────────────────────────────────────────────────

function exportToOpenTUI(node: ComponentNode): string {
  const imports = new Set<string>();
  const code = generateOpenTUIComponent(node, imports, 0);

  return `import { ${Array.from(imports).join(', ')} } from '@opentui/core';

function App() {
  return (
${code}
  );
}

export default App;
`;
}

function generateOpenTUIComponent(
  node: ComponentNode,
  imports: Set<string>,
  indent: number
): string {
  const spaces = '  '.repeat(indent + 1);
  const componentName = mapToOpenTUIComponent(node.type);
  imports.add(componentName);

  const props = generatePropsString(node);
  const style = generateStyleString(node);

  let result = `${spaces}<${componentName}${props}${style}`;

  if (node.children.length > 0) {
    result += '>\n';
    for (const child of node.children) {
      result += generateOpenTUIComponent(child, imports, indent + 1);
    }
    result += `${spaces}</${componentName}>\n`;
  } else {
    result += ' />\n';
  }

  return result;
}

// ── BubbleTea ─────────────────────────────────────────────────────────────────

function exportToBubbleTea(node: ComponentNode): string {
  return exportToBubbleTeaMain(node);
}

// ── Blessed ───────────────────────────────────────────────────────────────────

function exportToBlessed(node: ComponentNode): string {
  return `const blessed = require('blessed');

const screen = blessed.screen({
  smartCSR: true
});

${generateBlessedComponents(node, 0)}

screen.key(['escape', 'q', 'C-c'], function() {
  return process.exit(0);
});

screen.render();
`;
}

function generateBlessedComponents(node: ComponentNode, indent: number): string {
  const spaces = '  '.repeat(indent);
  const varName = node.name.replace(/[^a-zA-Z0-9]/g, '_').toLowerCase();

  const options: string[] = [];
  if (node.props.width) options.push(`width: ${JSON.stringify(node.props.width)}`);
  if (node.props.height) options.push(`height: ${JSON.stringify(node.props.height)}`);
  if (node.style.border) options.push(`border: { type: 'line' }`);
  if (node.props.content) options.push(`content: ${JSON.stringify(node.props.content)}`);

  let result = `${spaces}const ${varName} = blessed.box({\n`;
  result += options.map((opt) => `${spaces}  ${opt}`).join(',\n') + '\n';
  result += `${spaces}});\n`;
  result += `${spaces}screen.append(${varName});\n\n`;

  for (const child of node.children) {
    result += generateBlessedComponents(child, indent);
  }
  return result;
}

// ── Textual ───────────────────────────────────────────────────────────────────

function exportToTextual(node: ComponentNode): string {
  const imports = collectTextualImports(node);
  const css = generateTextualCss(node);
  const mountBody = generateTextualMountBody(node);
  const eventHandlers = generateTextualEventHandlers(imports.widgets);

  return `from textual.app import App, ComposeResult
${imports.containers.size ? `from textual.containers import ${Array.from(imports.containers).sort().join(', ')}\n` : ''}from textual.widgets import ${Array.from(imports.widgets).sort().join(', ')}


class MyApp(App):
${css ? `    CSS = ${pyString(css)}\n\n` : ''}    BINDINGS = [("q", "quit", "Quit")]

    def compose(self) -> ComposeResult:
${generateTextualComponents(node, 2) || '        pass\n'}
${mountBody ? `\n    def on_mount(self) -> None:\n${mountBody}` : ''}${eventHandlers}

if __name__ == "__main__":
    app = MyApp()
    app.run()
`;
}

interface TextualImports {
  containers: Set<string>;
  widgets: Set<string>;
}

interface TextualTable {
  id: string;
  columns: string[];
  rows: string[][];
}

interface TextualProgress {
  id: string;
  value: number;
}

function collectTextualImports(node: ComponentNode): TextualImports {
  const imports: TextualImports = {
    containers: new Set<string>(),
    widgets: new Set<string>(['Static']),
  };

  const collectImportsFromNode = (current: ComponentNode) => {
    if (current.hidden) return;

    const container = textualContainerFor(current);
    if (container) imports.containers.add(container);

    switch (current.type) {
      case 'Button':
        imports.widgets.add('Button');
        break;
      case 'TextInput':
        imports.widgets.add('Input');
        break;
      case 'Checkbox':
        imports.widgets.add('Checkbox');
        break;
      case 'Radio':
        imports.widgets.add('RadioButton');
        break;
      case 'Select':
        imports.widgets.add('Select');
        break;
      case 'Spinner':
        imports.widgets.add('LoadingIndicator');
        break;
      case 'ProgressBar':
        imports.widgets.add('ProgressBar');
        break;
      case 'Table':
        imports.widgets.add('DataTable');
        break;
      case 'List':
      case 'Menu':
      case 'Tabs':
        imports.widgets.add('Label');
        imports.widgets.add('ListItem');
        imports.widgets.add('ListView');
        break;
      default:
        break;
    }

    current.children.forEach(collectImportsFromNode);
  };

  collectImportsFromNode(node);
  return imports;
}

function generateTextualComponents(node: ComponentNode, indent: number): string {
  if (node.hidden) return '';

  const spaces = '  '.repeat(indent);
  const attrs = textualAttrs(node);
  const container = textualContainerFor(node);

  if (container) {
    const children = node.children
      .map((child) => generateTextualComponents(child, indent + 1))
      .join('');
    return `${spaces}with ${container}(${attrs}):\n${children || `${spaces}  yield Static(${pyString(node.name || node.type)})\n`}`;
  }

  switch (node.type) {
    case 'Text':
      return `${spaces}yield Static(${pyString((node.props.content as string) || 'Text')}${attrs ? `, ${attrs}` : ''})\n`;
    case 'Button':
      return `${spaces}yield Button(${pyString((node.props.label as string) || 'Button')}${attrs ? `, ${attrs}` : ''})\n`;
    case 'TextInput':
      return `${spaces}yield Input(value=${pyString((node.props.value as string) || '')}, placeholder=${pyString((node.props.placeholder as string) || '')}${attrs ? `, ${attrs}` : ''})\n`;
    case 'Checkbox':
      return `${spaces}yield Checkbox(${pyString((node.props.label as string) || 'Option')}, value=${pyBool(!!node.props.checked)}${attrs ? `, ${attrs}` : ''})\n`;
    case 'Radio':
      return `${spaces}yield RadioButton(${pyString((node.props.label as string) || 'Option')}, value=${pyBool(!!node.props.checked)}${attrs ? `, ${attrs}` : ''})\n`;
    case 'Select':
      return `${spaces}yield Select(${textualSelectOptions(node)}${attrs ? `, ${attrs}` : ''})\n`;
    case 'Spinner':
      return `${spaces}yield LoadingIndicator(${attrs})\n`;
    case 'ProgressBar': {
      const total = (node.props.max as number) ?? 100;
      return `${spaces}yield ProgressBar(total=${total}${attrs ? `, ${attrs}` : ''})\n`;
    }
    case 'Table':
      return `${spaces}yield DataTable(${attrs})\n`;
    case 'List':
    case 'Menu':
    case 'Tabs':
      return `${spaces}yield ListView(\n${textualListItems(node, indent + 1)}${spaces}${attrs})\n`;
    case 'Tree':
    case 'Breadcrumb':
      return `${spaces}yield Static(${pyString(textualStaticText(node))}${attrs ? `, ${attrs}` : ''})\n`;
    case 'Spacer':
      return `${spaces}yield Static("", ${attrs || 'classes="spacer"'})\n`;
    default:
      return `${spaces}yield Static(${pyString(node.name || node.type)}${attrs ? `, ${attrs}` : ''})\n`;
  }
}

function textualContainerFor(node: ComponentNode): string | null {
  if (node.type === 'Screen' || node.type === 'Box' || node.type === 'Modal') {
    return node.layout.direction === 'row' ? 'Horizontal' : 'Vertical';
  }

  if (node.type === 'Grid') return 'Grid';
  return null;
}

function textualAttrs(node: ComponentNode): string {
  const attrs: string[] = [];
  if (node.id) attrs.push(`id=${pyString(node.id)}`);

  const classes = [
    `tuistudio-${node.type.toLowerCase()}`,
    typeof node.props.className === 'string' ? node.props.className : '',
    typeof node.props.classes === 'string' ? node.props.classes : '',
  ]
    .filter(Boolean)
    .join(' ');

  if (classes) attrs.push(`classes=${pyString(classes)}`);
  if (node.type === 'Button' && node.props.disabled) attrs.push('disabled=True');
  return attrs.join(', ');
}

function textualSelectOptions(node: ComponentNode): string {
  const options = ((node.props.options as string[]) || ['Option 1', 'Option 2']).map((option) => [
    option,
    option.toLowerCase().replace(/\s+/g, '_'),
  ]);

  return `[${options.map(([label, value]) => `(${pyString(label)}, ${pyString(value)})`).join(', ')}]`;
}

function textualListItems(node: ComponentNode, indent: number): string {
  const spaces = '  '.repeat(indent);
  const labels =
    node.type === 'Tabs'
      ? ((node.props.tabs as any[]) || []).map((item) => textualItemLabel(item))
      : ((node.props.items as any[]) || []).map((item) => textualItemLabel(item));

  return labels.map((label) => `${spaces}ListItem(Label(${pyString(label)})),\n`).join('');
}

function textualStaticText(node: ComponentNode): string {
  if (node.type === 'Breadcrumb') {
    const separator = (node.props.separator as string) || ' / ';
    return ((node.props.items as any[]) || [])
      .map((item) => textualItemLabel(item))
      .join(separator);
  }

  if (node.type === 'Tree') {
    const lines: string[] = [];
    const renderTreeNode = (item: any, depth: number) => {
      const label = textualItemLabel(item);
      lines.push(`${'  '.repeat(depth)}${depth > 0 ? '└─ ' : ''}${label}`);
      ((item && item.children) || []).forEach((child: any) => renderTreeNode(child, depth + 1));
    };
    ((node.props.items as any[]) || []).forEach((item) => renderTreeNode(item, 0));
    return lines.join('\n') || 'Tree';
  }

  return node.name || node.type;
}

function textualItemLabel(item: any): string {
  if (typeof item === 'string') return item;
  const icon = item?.icon ? `${item.icon} ` : '';
  return `${icon}${item?.label || item?.name || 'Item'}`;
}

function generateTextualMountBody(node: ComponentNode): string {
  const tables = collectTextualTables(node);
  const progressBars = collectTextualProgressBars(node);
  if (!tables.length && !progressBars.length) return '';

  const tableBody = tables
    .map((table) => {
      const spaces = '        ';
      return (
        `${spaces}${textualVarName(table.id)} = self.query_one(${pyString(`#${table.id}`)}, DataTable)\n` +
        `${spaces}${textualVarName(table.id)}.add_columns(${table.columns.map(pyString).join(', ')})\n` +
        `${spaces}${textualVarName(table.id)}.add_rows(${pyTableRows(table.rows)})\n`
      );
    })
    .join('\n');

  const progressBody = progressBars
    .map((progress) => {
      const spaces = '        ';
      return `${spaces}${textualVarName(progress.id)} = self.query_one(${pyString(`#${progress.id}`)}, ProgressBar)\n${spaces}${textualVarName(progress.id)}.update(progress=${progress.value})\n`;
    })
    .join('\n');

  return [tableBody, progressBody].filter(Boolean).join('\n');
}

function collectTextualTables(node: ComponentNode): TextualTable[] {
  const tables: TextualTable[] = [];

  const visit = (current: ComponentNode) => {
    if (current.hidden) return;
    if (current.type === 'Table') {
      if (!current.id) return;
      tables.push({
        id: current.id,
        columns: (current.props.columns as string[]) || ['Column 1', 'Column 2'],
        rows: (current.props.rows as string[][]) || [],
      });
    }
    current.children.forEach(visit);
  };

  visit(node);
  return tables;
}

function collectTextualProgressBars(node: ComponentNode): TextualProgress[] {
  const progressBars: TextualProgress[] = [];

  const visit = (current: ComponentNode) => {
    if (current.hidden) return;
    if (current.type === 'ProgressBar') {
      if (!current.id) return;
      progressBars.push({
        id: current.id,
        value: (current.props.value as number) ?? 0,
      });
    }
    current.children.forEach(visit);
  };

  visit(node);
  return progressBars;
}

function generateTextualEventHandlers(widgets: Set<string>): string {
  const handlers: string[] = [];

  if (widgets.has('Button')) {
    handlers.push(`    def on_button_pressed(self, event: Button.Pressed) -> None:
        # Add button behavior here.
        pass
`);
  }

  if (widgets.has('Input')) {
    handlers.push(`    def on_input_changed(self, event: Input.Changed) -> None:
        # React to input changes here.
        pass
`);
  }

  if (widgets.has('Checkbox')) {
    handlers.push(`    def on_checkbox_changed(self, event: Checkbox.Changed) -> None:
        # React to checkbox changes here.
        pass
`);
  }

  if (widgets.has('Select')) {
    handlers.push(`    def on_select_changed(self, event: Select.Changed) -> None:
        # React to select changes here.
        pass
`);
  }

  return handlers.length ? `\n${handlers.join('\n')}` : '';
}

function generateTextualCss(node: ComponentNode): string {
  const rules: string[] = [];

  const visit = (current: ComponentNode) => {
    if (current.hidden) return;

    const declarations = textualCssDeclarations(current);
    if (current.id && declarations.length) {
      rules.push(`#${current.id} {\n${declarations.map((decl) => `  ${decl}`).join('\n')}\n}`);
    }

    current.children.forEach(visit);
  };

  visit(node);
  return rules.join('\n\n');
}

function textualCssDeclarations(node: ComponentNode): string[] {
  const declarations: string[] = [];
  const layout = node.layout as any;
  const style = node.style as any;

  if (typeof layout.width === 'number') declarations.push(`width: ${layout.width};`);
  if (typeof layout.height === 'number') declarations.push(`height: ${layout.height};`);
  if (layout.width === 'fill' || node.props.width === 'fill') declarations.push('width: 1fr;');
  if (layout.height === 'fill' || node.props.height === 'fill') declarations.push('height: 1fr;');
  if (layout.padding !== undefined)
    declarations.push(`padding: ${textualSpacing(layout.padding)};`);
  if (layout.margin !== undefined) declarations.push(`margin: ${textualSpacing(layout.margin)};`);
  if (layout.gap !== undefined && node.children.length)
    declarations.push(`grid-gutter: ${layout.gap};`);
  if (style.color) declarations.push(`color: ${style.color};`);
  if (style.backgroundColor) declarations.push(`background: ${style.backgroundColor};`);
  if (style.border)
    declarations.push(
      `border: ${textualBorderStyle(style.borderStyle)} ${style.borderColor || style.color || 'white'};`
    );
  const textStyles = [
    style.bold ? 'bold' : '',
    style.italic ? 'italic' : '',
    style.underline ? 'underline' : '',
  ].filter(Boolean);
  if (textStyles.length) declarations.push(`text-style: ${textStyles.join(' ')};`);

  return declarations;
}

function textualSpacing(
  value: number | { top: number; right: number; bottom: number; left: number }
): string {
  if (typeof value === 'number') return `${value}`;
  return `${value.top} ${value.right} ${value.bottom} ${value.left}`;
}

function textualBorderStyle(style?: string): string {
  const map: Record<string, string> = {
    single: 'solid',
    double: 'double',
    rounded: 'round',
    bold: 'heavy',
  };
  return map[style || 'single'] || 'solid';
}

function textualVarName(id: string): string {
  const safeName = id.replace(/[^a-zA-Z0-9_]/g, '_');
  return `widget_${safeName.length ? safeName : 'node'}`;
}

function pyTableRows(rows: string[][]): string {
  if (!rows.length) return '[]';
  return `[${rows.map(pyTuple).join(', ')}]`;
}

function pyTuple(row: string[]): string {
  const values = row.map((cell) => pyString(String(cell)));
  return `(${values.join(', ')}${values.length === 1 ? ',' : ''})`;
}

const PY_BACKSPACE = String.fromCharCode(8);

function pyString(value: string): string {
  // Escape common Python string literal sequences without using a control-character regex.
  return `"${value
    .replace(/\\/g, '\\\\')
    .replace(/"/g, '\\"')
    .replace(/\n/g, '\\n')
    .replace(/\r/g, '\\r')
    .replace(/\t/g, '\\t')
    .replace(/\f/g, '\\f')
    .split(PY_BACKSPACE)
    .join('\\b')}"`;
}

function pyBool(value: boolean): 'True' | 'False' {
  return value ? 'True' : 'False';
}

// ── Shared helpers ────────────────────────────────────────────────────────────

function mapToOpenTUIComponent(type: string): string {
  const map: Record<string, string> = {
    Box: 'Box',
    Text: 'Text',
    Button: 'Button',
    TextInput: 'Input',
  };
  return map[type] || 'Box';
}

function generatePropsString(node: ComponentNode): string {
  const props: string[] = [];
  if (node.props.width && typeof node.props.width === 'number')
    props.push(`width={${node.props.width}}`);
  if (node.props.height && typeof node.props.height === 'number')
    props.push(`height={${node.props.height}}`);
  return props.length ? ' ' + props.join(' ') : '';
}

function generateStyleString(node: ComponentNode): string {
  const styles: string[] = [];
  if (node.style.color) styles.push(`color="${node.style.color}"`);
  if (node.style.backgroundColor) styles.push(`backgroundColor="${node.style.backgroundColor}"`);
  if (node.style.border) styles.push(`border={true}`);
  if (node.style.bold) styles.push(`bold={true}`);
  return styles.length ? ' ' + styles.join(' ') : '';
}
