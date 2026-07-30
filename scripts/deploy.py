#!/usr/bin/env python3
import os
import shutil
import re
from pathlib import Path

SOURCE_DIR = Path("/home/d.b.hicks/go-app")
DIST_DIR = SOURCE_DIR / "dist"
DOCS_DIR = SOURCE_DIR / "docs"
WWW_DIR = Path("/home/d.b.hicks/www/d.b.hicks/go-app")
README_PATH = SOURCE_DIR / "README.md"
INDEX_PATH = WWW_DIR / "index.html"

EXECUTABLES = [
    {
        "filename": "go-app-linux-amd64",
        "name": "Linux (x64)",
        "icon": "🐧",
        "desc": "64-bit Linux Desktop / Server"
    },
    {
        "filename": "go-app-windows-amd64.exe",
        "name": "Windows (x64)",
        "icon": "🪟",
        "desc": "64-bit Windows Desktop"
    },
    {
        "filename": "go-app-darwin-amd64",
        "name": "macOS (Intel x64)",
        "icon": "🍎",
        "desc": "Intel-based Mac computers"
    },
    {
        "filename": "go-app-darwin-arm64",
        "name": "macOS (Apple Silicon)",
        "icon": "🍏",
        "desc": "Apple M1 / M2 / M3 / M4 Macs"
    },
    {
        "filename": "go-app-rpi-arm64",
        "name": "Raspberry Pi (ARM64)",
        "icon": "🍓",
        "desc": "Raspberry Pi 3/4/5 (64-bit OS)"
    },
    {
        "filename": "go-app-rpi-armv7",
        "name": "Raspberry Pi (ARMv7 32-bit)",
        "icon": "🍓",
        "desc": "Raspberry Pi 2/3/4 (32-bit OS)"
    },
]

def format_size(bytes_size):
    for unit in ['B', 'KB', 'MB', 'GB']:
        if bytes_size < 1024.0:
            return f"{bytes_size:.1f} {unit}"
        bytes_size /= 1024.0
    return f"{bytes_size:.1f} TB"

def markdown_to_html(md_text):
    lines = md_text.split('\n')
    html_lines = []
    in_code_block = False
    in_table = False
    table_lines = []

    def render_table(t_lines):
        if not t_lines:
            return ""
        out = ["<div class='table-container'><table>"]
        for i, line in enumerate(t_lines):
            cells = [c.strip() for c in line.strip('|').split('|')]
            if i == 1 and all(set(c.strip()) <= {'-', ':'} for c in cells):
                continue
            tag = "th" if i == 0 else "td"
            out.append("<tr>" + "".join(f"<{tag}>{c}</{tag}>" for c in cells) + "</tr>")
        out.append("</table></div>")
        return "\n".join(out)

    for line in lines:
        if line.startswith("```"):
            if in_code_block:
                html_lines.append("</code></pre>")
                in_code_block = False
            else:
                lang = line[3:].strip()
                html_lines.append(f"<pre><code class='language-{lang}'>")
                in_code_block = True
            continue

        if in_code_block:
            escaped = line.replace('&', '&amp;').replace('<', '&lt;').replace('>', '&gt;')
            html_lines.append(escaped)
            continue

        if line.startswith("|") and line.endswith("|"):
            if not in_table:
                in_table = True
                table_lines = []
            table_lines.append(line)
            continue
        elif in_table:
            html_lines.append(render_table(table_lines))
            in_table = False
            table_lines = []

        if line.startswith("# "):
            html_lines.append(f"<h1>{line[2:]}</h1>")
        elif line.startswith("## "):
            html_lines.append(f"<h2>{line[3:]}</h2>")
        elif line.startswith("### "):
            html_lines.append(f"<h3>{line[4:]}</h3>")
        elif line.startswith("---"):
            html_lines.append("<hr/>")
        elif line.startswith("- "):
            html_lines.append(f"<li>{line[2:]}</li>")
        elif line.strip() == "":
            html_lines.append("<p></p>")
        else:
            processed = line
            processed = re.sub(r'\*\*(.*?)\*\*', r'<strong>\1</strong>', processed)
            processed = re.sub(r'`(.*?)`', r'<code>\1</code>', processed)
            processed = re.sub(r'\[(.*?)\]\((.*?)\)', r'<a href="\2" target="_blank">\1</a>', processed)
            html_lines.append(f"<p>{processed}</p>")

    if in_table:
        html_lines.append(render_table(table_lines))

    return "\n".join(html_lines)

def generate_html(download_prefix):
    downloads_html_cards = []

    for item in EXECUTABLES:
        size_str = format_size((DIST_DIR / item["filename"]).stat().st_size) if (DIST_DIR / item["filename"]).exists() else "Unavailable"

        downloads_html_cards.append(f"""
        <div class="download-card">
            <div class="card-icon">{item['icon']}</div>
            <div class="card-info">
                <h3>{item['name']}</h3>
                <p>{item['desc']}</p>
                <div class="file-meta">File: <code>{item['filename']}</code> • {size_str}</div>
            </div>
            <a href="{download_prefix}{item['filename']}" class="btn-download" download>
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                Download
            </a>
        </div>
        """)

    readme_content = README_PATH.read_text(encoding="utf-8") if README_PATH.exists() else ""
    body_html = markdown_to_html(readme_content)

    return f"""<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Application Distribution & Documentation</title>

    <style>
        :root {{
            --bg-color: #0f172a;
            --card-bg: #1e293b;
            --border-color: #334155;
            --text-primary: #f8fafc;
            --text-secondary: #94a3b8;
            --accent-color: #38bdf8;
            --accent-hover: #0284c7;
            --code-bg: #090d16;
        }}
        * {{ box-sizing: border-box; margin: 0; padding: 0; }}
        body {{
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background-color: var(--bg-color);
            color: var(--text-primary);
            line-height: 1.6;
            padding: 2rem 1rem;
        }}
        .container {{
            max-width: 1000px;
            margin: 0 auto;
        }}
        header {{
            text-align: center;
            margin-bottom: 3rem;
            padding-bottom: 2rem;
            border-bottom: 1px solid var(--border-color);
        }}
        header h1 {{
            font-size: 2.5rem;
            font-weight: 700;
            background: linear-gradient(135deg, #38bdf8, #818cf8);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 0.5rem;
        }}
        header p {{
            color: var(--text-secondary);
            font-size: 1.1rem;
        }}
        .section-title {{
            font-size: 1.8rem;
            margin: 2.5rem 0 1.5rem 0;
            color: var(--text-primary);
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }}
        .downloads-grid {{
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 1.25rem;
            margin-bottom: 3rem;
        }}
        .download-card {{
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            border-radius: 12px;
            padding: 1.25rem;
            display: flex;
            flex-direction: column;
            justify-content: space-between;
            transition: transform 0.2s ease, border-color 0.2s ease;
        }}
        .download-card:hover {{
            transform: translateY(-2px);
            border-color: var(--accent-color);
        }}
        .card-icon {{
            font-size: 2rem;
            margin-bottom: 0.5rem;
        }}
        .card-info h3 {{
            font-size: 1.2rem;
            font-weight: 600;
            margin-bottom: 0.25rem;
        }}
        .card-info p {{
            color: var(--text-secondary);
            font-size: 0.9rem;
            margin-bottom: 0.75rem;
        }}
        .file-meta {{
            font-size: 0.8rem;
            color: var(--text-secondary);
            margin-bottom: 1rem;
        }}
        .btn-download {{
            display: inline-flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
            background: var(--accent-color);
            color: #0f172a;
            font-weight: 600;
            padding: 0.6rem 1rem;
            border-radius: 8px;
            text-decoration: none;
            transition: background 0.2s ease;
        }}
        .btn-download:hover {{
            background: var(--accent-hover);
            color: #ffffff;
        }}
        .docs-section {{
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            border-radius: 12px;
            padding: 2.5rem;
            margin-top: 2rem;
        }}
        .docs-section h1 {{ font-size: 2rem; margin-bottom: 1rem; color: var(--accent-color); }}
        .docs-section h2 {{ font-size: 1.5rem; margin: 1.5rem 0 1rem 0; border-bottom: 1px solid var(--border-color); padding-bottom: 0.5rem; }}
        .docs-section h3 {{ font-size: 1.2rem; margin: 1.2rem 0 0.5rem 0; }}
        .docs-section p {{ margin-bottom: 1rem; color: #cbd5e1; }}
        .docs-section ul {{ margin-left: 1.5rem; margin-bottom: 1rem; color: #cbd5e1; }}
        .docs-section li {{ margin-bottom: 0.4rem; }}
        pre {{
            background: var(--code-bg);
            padding: 1rem;
            border-radius: 8px;
            overflow-x: auto;
            border: 1px solid var(--border-color);
            margin-bottom: 1.2rem;
        }}
        code {{
            font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
            font-size: 0.9rem;
            color: #e2e8f0;
        }}
        p code, li code {{
            background: var(--code-bg);
            padding: 0.2rem 0.4rem;
            border-radius: 4px;
            color: var(--accent-color);
        }}
        .table-container {{
            overflow-x: auto;
            margin-bottom: 1.5rem;
        }}
        table {{
            width: 100%;
            border-collapse: collapse;
            margin: 1rem 0;
        }}
        th, td {{
            padding: 0.75rem 1rem;
            text-align: left;
            border-bottom: 1px solid var(--border-color);
        }}
        th {{
            background: rgba(255, 255, 255, 0.05);
            color: var(--accent-color);
            font-weight: 600;
        }}
        a {{ color: var(--accent-color); text-decoration: none; }}
        a:hover {{ text-decoration: underline; }}
        .github-link {{
            display: inline-block;
            margin-top: 0.5rem;
            padding: 0.4rem 1rem;
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            font-size: 0.9rem;
            font-weight: 500;
            transition: border-color 0.2s ease;
        }}
        .github-link:hover {{
            border-color: var(--accent-color);
            text-decoration: none;
        }}
        hr {{ border: none; border-top: 1px solid var(--border-color); margin: 2rem 0; }}
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>⚡ Go Application Release Distribution</h1>
            <p>Self-contained, cross-platform executable builds with embedded React UI and OpenAPI docs.</p>
            <p><a href="https://github.com/dhicks6345789/go-app" target="_blank" class="github-link">View on GitHub</a> &middot; <a href="./docs/api.html" target="_blank" class="github-link">API Documentation</a></p>
        </header>

        <h2 class="section-title">📦 Download Executables</h2>
        <div class="downloads-grid">
            {''.join(downloads_html_cards)}
        </div>

        <h2 class="section-title">📖 Project Documentation</h2>
        <div class="docs-section">
            {body_html}
        </div>
    </div>
</body>
</html>
"""


def main():
    WWW_DIR.mkdir(parents=True, exist_ok=True)
    print(f"Target distribution directory: {WWW_DIR}")

    for item in EXECUTABLES:
        src_path = DIST_DIR / item["filename"]
        dest_path = WWW_DIR / item["filename"]

        if src_path.exists():
            shutil.copy2(src_path, dest_path)
            print(f"Copied {item['filename']} ({format_size(dest_path.stat().st_size)}) -> {dest_path}")
        else:
            print(f"Warning: {src_path} not found. Skipping file copy.")

    # Copy swaggo-generated API documentation
    www_docs = WWW_DIR / "docs"
    www_docs.mkdir(parents=True, exist_ok=True)
    for doc_file in DOCS_DIR.iterdir():
        if doc_file.is_file() and doc_file.suffix in (".json", ".yaml"):
            shutil.copy2(doc_file, www_docs / doc_file.name)
            print(f"Copied docs/{doc_file.name} -> {www_docs / doc_file.name}")

    # Copy API docs HTML page
    api_template = SOURCE_DIR / "scripts" / "api_docs_template.html"
    if api_template.exists():
        shutil.copy2(api_template, www_docs / "api.html")
        print(f"Copied api_docs_template.html -> {www_docs / 'api.html'}")
    else:
        print("Warning: api_docs_template.html not found")

    www_html = generate_html("./")
    INDEX_PATH.write_text(www_html, encoding="utf-8")
    print(f"Generated distribution page at: {INDEX_PATH}")

    repo_html = generate_html("./dist/")
    REPO_INDEX = SOURCE_DIR / "index.html"
    REPO_INDEX.write_text(repo_html, encoding="utf-8")
    print(f"Generated repository index.html at: {REPO_INDEX}")

if __name__ == "__main__":
    main()
