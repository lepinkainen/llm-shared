#!/usr/bin/env python3
"""
Python function analyzer - similar to gofuncs.go but for Python codebases.
Extracts function information in LLM-optimized format.
"""

import argparse
import ast
import os
import sys
from dataclasses import dataclass
from typing import Optional


@dataclass
class FunctionInfo:
    file: str
    line: int
    type: str  # f=function, m=method, s=staticmethod, c=classmethod, p=property
    exported: bool  # True if public (not starting with _)
    name: str
    class_name: Optional[str]
    params: list[str]
    returns: Optional[str]
    decorators: list[str]
    is_async: bool


class PythonFunctionExtractor(ast.NodeVisitor):
    def __init__(self, file_path: str, relative_path: str):
        self.file_path = file_path
        self.relative_path = relative_path
        self.functions: list[FunctionInfo] = []
        self.class_stack: list[str] = []

    def visit_ClassDef(self, node: ast.ClassDef):
        self.class_stack.append(node.name)
        self.generic_visit(node)
        self.class_stack.pop()

    def visit_FunctionDef(self, node: ast.FunctionDef):
        self._process_function(node, is_async=False)

    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef):
        self._process_function(node, is_async=True)

    def _process_function(self, node, is_async: bool):
        # Determine function type and class context
        class_name = self.class_stack[-1] if self.class_stack else None
        func_type = self._determine_function_type(node, class_name)

        # Extract decorators
        decorators = [self._get_decorator_name(dec) for dec in node.decorator_list]

        # Extract parameters
        params = self._extract_parameters(node.args)

        # Extract return type
        returns = (
            self._extract_return_type(node.returns)
            if hasattr(node, "returns") and node.returns
            else None
        )

        # Determine if exported (public)
        exported = not node.name.startswith("_")

        func_info = FunctionInfo(
            file=self.relative_path,
            line=node.lineno,
            type=func_type,
            exported=exported,
            name=node.name,
            class_name=class_name,
            params=params,
            returns=returns,
            decorators=decorators,
            is_async=is_async,
        )

        self.functions.append(func_info)

    def _determine_function_type(self, node, class_name: Optional[str]) -> str:
        if not class_name:
            return "f"  # function

        # Check decorators for special method types
        for decorator in node.decorator_list:
            dec_name = self._get_decorator_name(decorator)
            if dec_name == "staticmethod":
                return "s"
            elif dec_name == "classmethod":
                return "c"
            elif dec_name == "property":
                return "p"

        return "m"  # method

    def _get_decorator_name(self, decorator) -> str:
        if isinstance(decorator, ast.Name):
            return decorator.id
        elif isinstance(decorator, ast.Attribute):
            return decorator.attr
        elif isinstance(decorator, ast.Call):
            if isinstance(decorator.func, ast.Name):
                return decorator.func.id
            elif isinstance(decorator.func, ast.Attribute):
                return decorator.func.attr
        return "unknown"

    def _extract_parameters(self, args: ast.arguments) -> list[str]:
        params = []

        # Regular arguments
        for arg in args.args:
            param_str = arg.arg
            if arg.annotation:
                param_str += ":" + self._extract_type_annotation(arg.annotation)
            params.append(param_str)

        # *args
        if args.vararg:
            vararg_str = "*" + args.vararg.arg
            if args.vararg.annotation:
                vararg_str += ":" + self._extract_type_annotation(
                    args.vararg.annotation
                )
            params.append(vararg_str)

        # **kwargs
        if args.kwarg:
            kwarg_str = "**" + args.kwarg.arg
            if args.kwarg.annotation:
                kwarg_str += ":" + self._extract_type_annotation(args.kwarg.annotation)
            params.append(kwarg_str)

        # Keyword-only arguments
        for arg in args.kwonlyargs:
            param_str = arg.arg
            if arg.annotation:
                param_str += ":" + self._extract_type_annotation(arg.annotation)
            params.append(param_str)

        return params

    def _extract_return_type(self, returns) -> str:
        return self._extract_type_annotation(returns)

    def _extract_type_annotation(self, annotation) -> str:
        if isinstance(annotation, ast.Name):
            return annotation.id
        elif isinstance(annotation, ast.Constant):
            return repr(annotation.value)
        elif isinstance(annotation, ast.Attribute):
            return (
                f"{self._extract_type_annotation(annotation.value)}.{annotation.attr}"
            )
        elif isinstance(annotation, ast.Subscript):
            value = self._extract_type_annotation(annotation.value)
            slice_val = self._extract_type_annotation(annotation.slice)
            return f"{value}[{slice_val}]"
        elif isinstance(annotation, ast.Tuple):
            elements = [self._extract_type_annotation(elt) for elt in annotation.elts]
            return f"({','.join(elements)})"
        elif isinstance(annotation, ast.List):
            elements = [self._extract_type_annotation(elt) for elt in annotation.elts]
            return f"[{','.join(elements)}]"
        else:
            return "unknown"


def extract_functions(directory: str) -> list[FunctionInfo]:
    """Extract function information from all Python files in directory."""
    functions = []

    for root, dirs, files in os.walk(directory):
        # Skip common non-source directories
        dirs[:] = [
            d
            for d in dirs
            if not d.startswith(".") and d not in {"__pycache__", "node_modules"}
        ]

        for file in files:
            if not file.endswith(".py"):
                continue

            file_path = os.path.join(root, file)
            relative_path = os.path.relpath(file_path, directory)

            try:
                with open(file_path, encoding="utf-8") as f:
                    source = f.read()

                tree = ast.parse(source, filename=file_path)
                extractor = PythonFunctionExtractor(file_path, relative_path)
                extractor.visit(tree)
                functions.extend(extractor.functions)

            except (SyntaxError, UnicodeDecodeError) as e:
                print(f"Warning: Could not parse {relative_path}: {e}", file=sys.stderr)
                continue

    return functions


def format_function(fn: FunctionInfo) -> str:
    """Format function info similar to gofuncs output."""
    exported = "y" if fn.exported else "n"

    # Build signature
    signature = f"({','.join(fn.params)})"
    if fn.returns:
        signature += f"->{fn.returns}"

    # Add async prefix if needed
    if fn.is_async:
        signature = f"async {signature}"

    # Add decorators if any
    decorators_str = ",".join(fn.decorators) if fn.decorators else ""

    # Format based on whether it's a class method or function
    if fn.class_name:
        return f"{fn.file}:{fn.line}:{fn.type}:{exported}:{fn.name}:{fn.class_name}:{signature}:{decorators_str}"
    else:
        return f"{fn.file}:{fn.line}:{fn.type}:{exported}:{fn.name}::{signature}:{decorators_str}"


def main():
    parser = argparse.ArgumentParser(description="Extract Python function information")
    parser.add_argument("--dir", default=".", help="Directory to scan for Python files")
    args = parser.parse_args()

    try:
        functions = extract_functions(args.dir)

        # Sort by file, then by line number (same as gofuncs)
        functions.sort(key=lambda f: (f.file, f.line))

        for fn in functions:
            print(format_function(fn))

    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
