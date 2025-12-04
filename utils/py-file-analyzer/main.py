#!/usr/bin/env python3
"""
Python File Analyzer - Similar to go-file-analyzer but for Python codebases.
Lists file sizes, function lengths, and complexity metrics to identify refactoring opportunities.
"""

import argparse
import ast
import os
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Optional, Set, Tuple


@dataclass
class FunctionInfo:
    name: str
    lines: int
    statements: int
    is_async: bool
    is_method: bool
    class_name: Optional[str] = None


@dataclass
class FileAnalysis:
    path: str
    lines: int
    classes: List[str]
    functions: List[FunctionInfo]
    imports: List[str]
    method_counts: Dict[str, int]
    top_level_functions: int
    notes: List[str]


class ComplexityAnalyzer(ast.NodeVisitor):
    def __init__(self, file_path: str, source_lines: List[str]):
        self.file_path = file_path
        self.source_lines = source_lines
        self.functions: List[FunctionInfo] = []
        self.classes: List[str] = []
        self.imports: List[str] = []
        self.class_stack: List[str] = []
        
    def visit_Import(self, node: ast.Import):
        for alias in node.names:
            self.imports.append(alias.name)
        self.generic_visit(node)
        
    def visit_ImportFrom(self, node: ast.ImportFrom):
        module = node.module or ""
        for alias in node.names:
            if module:
                self.imports.append(f"{module}.{alias.name}")
            else:
                self.imports.append(alias.name)
        self.generic_visit(node)
    
    def visit_ClassDef(self, node: ast.ClassDef):
        self.classes.append(node.name)
        self.class_stack.append(node.name)
        self.generic_visit(node)
        self.class_stack.pop()
    
    def visit_FunctionDef(self, node: ast.FunctionDef):
        self._process_function(node, is_async=False)
        
    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef):
        self._process_function(node, is_async=True)
    
    def _process_function(self, node, is_async: bool):
        # Determine function boundaries
        start_line = node.lineno
        end_line = self._find_function_end(node)
        lines = end_line - start_line + 1
        
        # Count statements (approximate by AST node count in body)
        statements = len(node.body)
        
        # Determine if it's a method
        is_method = bool(self.class_stack)
        class_name = self.class_stack[-1] if self.class_stack else None
        
        func_info = FunctionInfo(
            name=node.name,
            lines=lines,
            statements=statements,
            is_async=is_async,
            is_method=is_method,
            class_name=class_name
        )
        
        self.functions.append(func_info)
        self.generic_visit(node)
    
    def _find_function_end(self, node) -> int:
        """Find the end line of a function by looking at the last node."""
        def find_last_node(n):
            if hasattr(n, 'lineno'):
                last = n.lineno
            else:
                last = 0
            
            if hasattr(n, 'end_lineno') and n.end_lineno:
                return max(last, n.end_lineno)
            
            # For child nodes, find the one with the highest line number
            for child in ast.walk(n):
                if hasattr(child, 'lineno'):
                    last = max(last, child.lineno)
                if hasattr(child, 'end_lineno') and child.end_lineno:
                    last = max(last, child.end_lineno)
            return last
        
        end = find_last_node(node)
        # Handle case where function has no content
        if end < node.lineno:
            end = node.lineno
        return end


def count_file_lines(file_path: str) -> int:
    """Count total lines in a Python file."""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            return len(f.readlines())
    except (UnicodeDecodeError, IOError):
        return 0


def analyze_file(file_path: str, root_dir: str) -> FileAnalysis:
    """Analyze a Python file and return detailed metrics."""
    full_path = os.path.join(root_dir, file_path) if not os.path.isabs(file_path) else file_path
    lines = count_file_lines(full_path)
    if lines == 0:
        return FileAnalysis(
            path=file_path,
            lines=0,
            classes=[],
            functions=[],
            imports=[],
            method_counts={},
            top_level_functions=0,
            notes=["Could not read or decode file"]
        )
    
    try:
        with open(full_path, 'r', encoding='utf-8') as f:
            source = f.read()
    except (UnicodeDecodeError, IOError) as e:
        return FileAnalysis(
            path=file_path,
            lines=lines,
            classes=[],
            functions=[],
            imports=[],
            method_counts={},
            top_level_functions=0,
            notes=[f"Could not read file: {e}"]
        )
    
    source_lines = source.splitlines()
    analyzer = ComplexityAnalyzer(file_path, source_lines)
    
    try:
        tree = ast.parse(source, filename=file_path)
        analyzer.visit(tree)
    except SyntaxError as e:
        return FileAnalysis(
            path=file_path,
            lines=lines,
            classes=[],
            functions=[],
            imports=[],
            method_counts={},
            top_level_functions=0,
            notes=[f"Syntax error: {e}"]
        )
    
    # Calculate method counts per class
    method_counts = {}
    top_level_functions = 0
    
    for func in analyzer.functions:
        if func.is_method and func.class_name:
            method_counts[func.class_name] = method_counts.get(func.class_name, 0) + 1
        elif not func.is_method:
            top_level_functions += 1
    
    # Generate notes for complexity concerns
    notes = []
    
    # Too many classes
    if len(analyzer.classes) > 3:
        notes.append(f"Many classes in file (classes={len(analyzer.classes)})")
    
    # Long functions
    long_funcs = [f for f in analyzer.functions if f.lines > 50]
    very_long_funcs = [f for f in analyzer.functions if f.lines > 100]
    
    if very_long_funcs:
        notes.append(f"Very long functions: {len(very_long_funcs)} >100 lines")
    elif long_funcs:
        notes.append(f"Long functions: {len(long_funcs)} >50 lines")
    
    # High statement count
    high_stmt_funcs = [f for f in analyzer.functions if f.statements > 20]
    if high_stmt_funcs:
        notes.append(f"Complex functions: {len(high_stmt_funcs)} >20 statements")
    
    # Mixed concerns (both async and sync functions)
    has_async = any(f.is_async for f in analyzer.functions)
    has_sync = any(not f.is_async for f in analyzer.functions)
    if has_async and has_sync and len(analyzer.functions) > 5:
        notes.append("Mixed async/sync patterns")
    
    # Import patterns
    web_imports = ['flask', 'django', 'fastapi', 'tornado', 'aiohttp', 'requests', 'httpx']
    db_imports = ['sqlalchemy', 'psycopg2', 'pymongo', 'sqlite3', 'redis', 'asyncpg']
    
    has_web = any(any(imp.startswith(web) for web in web_imports) for imp in analyzer.imports)
    has_db = any(any(imp.startswith(db) for db in db_imports) for imp in analyzer.imports)
    
    if has_web and has_db:
        notes.append("Mixed web and database concerns")
    
    return FileAnalysis(
        path=file_path,
        lines=lines,
        classes=analyzer.classes,
        functions=analyzer.functions,
        imports=analyzer.imports,
        method_counts=method_counts,
        top_level_functions=top_level_functions,
        notes=notes
    )


def collect_python_files(root_dir: str, exclude_tests: bool = True) -> List[FileAnalysis]:
    """Collect and analyze all Python files in a directory."""
    analyses = []
    skip_dirs = {'.git', '__pycache__', '.pytest_cache', '.tox', 'venv', 'env', '.venv', 
                'node_modules', 'build', 'dist', 'target', 'vendor'}
    
    for root, dirs, files in os.walk(root_dir):
        # Skip common non-source directories
        dirs[:] = [d for d in dirs if d not in skip_dirs]
        
        for file in files:
            if file.endswith('.py') and not file.endswith('_pb2.py'):  # Skip protobuf generated files
                # Skip __init__.py files (package structure files)
                if file == '__init__.py':
                    continue
                    
                # Skip test files if exclude_tests is True
                if exclude_tests and (file.startswith('test_') or file.endswith('_test.py')):
                    continue
                    
                file_path = os.path.join(root, file)
                rel_path = os.path.relpath(file_path, root_dir)
                analysis = analyze_file(rel_path, root_dir)
                analyses.append(analysis)
    
    return analyses


def print_analysis(analysis: FileAnalysis, top_functions: int = 5):
    """Print analysis for a single file in a compact format."""
    print(f"{analysis.lines:8d} {analysis.path}")
    
    # Classes
    if analysis.classes:
        print(f"    classes: {len(analysis.classes)} ({', '.join(analysis.classes[:3])}{'...' if len(analysis.classes) > 3 else ''})")
    else:
        print("    classes: 0")
    
    # Methods
    if analysis.method_counts or analysis.top_level_functions:
        method_parts = []
        for class_name, count in sorted(analysis.method_counts.items()):
            method_parts.append(f"{class_name}={count}")
        if analysis.top_level_functions > 0:
            method_parts.append(f"top-level={analysis.top_level_functions}")
        print(f"    methods: {'; '.join(method_parts)}")
    else:
        print("    methods: none")
    
    # Line length buckets
    buckets = _calculate_function_buckets(analysis.functions)
    print(f"    buckets: >100={buckets['over100']}, 50-100={buckets['between50_100']}, 20-49={buckets['between20_49']}, <20={buckets['under20']}")
    
    # Top functions by lines
    if analysis.functions:
        sorted_funcs = sorted(analysis.functions, key=lambda f: f.lines, reverse=True)
        limit = min(top_functions, len(sorted_funcs))
        print(f"    funcs (top {limit} by lines):")
        for i, func in enumerate(sorted_funcs[:limit]):
            async_prefix = "async " if func.is_async else ""
            method_prefix = f"{func.class_name}." if func.class_name else ""
            print(f"      {func.lines:4d} lines | stmts={func.statements} | {async_prefix}{method_prefix}{func.name}")
    
    # Notes
    for note in analysis.notes:
        print(f"    note: {note}")


def _calculate_function_buckets(functions: List[FunctionInfo]) -> Dict[str, int]:
    """Calculate function length buckets."""
    buckets = {
        'over100': 0,
        'between50_100': 0,
        'between20_49': 0,
        'under20': 0
    }
    
    for func in functions:
        if func.lines > 100:
            buckets['over100'] += 1
        elif func.lines >= 50:
            buckets['between50_100'] += 1
        elif func.lines >= 20:
            buckets['between20_49'] += 1
        else:
            buckets['under20'] += 1
    
    return buckets


def main():
    parser = argparse.ArgumentParser(
        description="Analyze Python files for complexity and size",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s -dir .                    # Analyze current directory (excluding tests)
  %(prog)s -dir . -n 10 -topfuncs 3 # Show top 10 files with 3 functions each
  %(prog)s -dir . --include-tests   # Include test files in analysis
        """
    )
    
    parser.add_argument('-dir', default='.', 
                       help='Directory to scan for Python files (default: current directory)')
    parser.add_argument('-n', type=int, default=20,
                       help='Number of files to display (default: 20)')
    parser.add_argument('-topfuncs', type=int, default=5,
                       help='Number of functions to list per file (default: 5)')
    parser.add_argument('--include-tests', action='store_true',
                       help='Include test files (test_*.py and *_test.py) in analysis')
    
    args = parser.parse_args()
    
    try:
        exclude_tests = not args.include_tests
        analyses = collect_python_files(args.dir, exclude_tests)
        
        if not analyses:
            print("No Python files found")
            return
        
        # Sort by line count (descending)
        analyses.sort(key=lambda a: a.lines, reverse=True)
        
        # Limit results
        limit = min(args.n, len(analyses))
        
        for i in range(limit):
            if i > 0:
                print()  # Blank line between files
            print_analysis(analyses[i], args.topfuncs)
        
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()