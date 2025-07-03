#!/usr/bin/env node
/**
 * JavaScript/TypeScript function analyzer - similar to gofuncs.go and pyfuncs.py
 * Extracts function information in LLM-optimized format.
 */

const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

class FunctionInfo {
  constructor({
    file,
    line,
    type,
    exported,
    name,
    className = null,
    params = [],
    returns = null,
    decorators = [],
    isAsync = false,
    isGenerator = false,
  }) {
    this.file = file;
    this.line = line;
    this.type = type; // f=function, m=method, a=arrow, c=constructor, g=getter, s=setter
    this.exported = exported;
    this.name = name;
    this.className = className;
    this.params = params;
    this.returns = returns;
    this.decorators = decorators;
    this.isAsync = isAsync;
    this.isGenerator = isGenerator;
  }
}

class JSFunctionExtractor {
  constructor(filePath, relativePath) {
    this.filePath = filePath;
    this.relativePath = relativePath;
    this.functions = [];
    this.classStack = [];
    this.content = "";
    this.lines = [];
  }

  extract() {
    try {
      this.content = fs.readFileSync(this.filePath, "utf8");
      this.lines = this.content.split("\n");

      // Try to use TypeScript parser if available, otherwise use regex
      if (this.hasTypeScript()) {
        this.extractWithTypeScript();
      } else {
        this.extractWithRegex();
      }
    } catch (error) {
      console.error(
        `Warning: Could not parse ${this.relativePath}: ${error.message}`
      );
    }

    return this.functions;
  }

  hasTypeScript() {
    try {
      require("@typescript-eslint/parser");
      return true;
    } catch {
      return false;
    }
  }

  extractWithTypeScript() {
    // TypeScript AST parsing is not implemented; falling back to regex extraction.
    this.extractWithRegex();
  }

  extractWithRegex() {
    // Match various function patterns
    const patterns = [
      // Regular functions: function name() {}
      /(?:^|\s+)(export\s+)?(?:async\s+)?function\s*\*?\s+(\w+)\s*\(([^)]*)\)(?:\s*:\s*([^{]+?))?\s*\{/gm,

      // Arrow functions: const name = () => {}
      /(?:^|\s+)(export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?\(([^)]*)\)(?:\s*:\s*([^=]+?))?\s*=>/gm,
      // Matches: const name = (params) => ... and const name = param => ...
      /(?:^|\s+)(export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?(?:\(([^)]*)\)|(\w+))(?:\s*:\s*([^=]+?))?\s*=>/gm,
      // Class methods: methodName() {}
      /(?:^|\s+)((?:public|private|protected|static)\s+)?(?:async\s+)?(?:get\s+|set\s+)?(\w+)\s*\(([^)]*)\)(?:\s*:\s*([^{]+?))?\s*\{/gm,

      // Object methods: name: function() {} or name() {}
      /(\w+)\s*:\s*(?:async\s+)?function\s*\*?\s*\(([^)]*)\)(?:\s*:\s*([^{]+?))?\s*\{/gm,
    ];

    patterns.forEach((pattern) => {
      let match;
      while ((match = pattern.exec(this.content)) !== null) {
        this.processMatch(match);
      }
    });

    // Reset regex lastIndex
    patterns.forEach((pattern) => (pattern.lastIndex = 0));
  }

  processMatch(match) {
    const lineNumber = this.getLineNumber(match.index);
    const fullMatch = match[0];

    let exported = false;
    let isAsync = false;
    let isGenerator = false;
    let type = "f"; // function
    let name = "";
    let params = [];
    let returnType = null;

    // Determine function type and extract details
    if (fullMatch.includes("export")) {
      exported = true;
    }

    if (fullMatch.includes("async")) {
      isAsync = true;
    }

    if (fullMatch.includes("function*") || fullMatch.includes("*")) {
      isGenerator = true;
    }

    // Extract function name and parameters based on pattern
    if (fullMatch.includes("function")) {
      // Regular function or function expression
      const functionMatch = fullMatch.match(
        /function\s*\*?\s+(\w+)\s*\(([^)]*)\)(?:\s*:\s*([^{]+?))?/
      );
      if (functionMatch) {
        name = functionMatch[1];
        params = this.parseParameters(functionMatch[2]);
        returnType = functionMatch[3]?.trim();
      }
    } else if (fullMatch.includes("=>")) {
      // Arrow function
      const arrowMatch = fullMatch.match(
        /(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?\(([^)]*)\)(?:\s*:\s*([^=]+?))?\s*=>/
      );
      if (arrowMatch) {
        name = arrowMatch[1];
        params = this.parseParameters(arrowMatch[2]);
        returnType = arrowMatch[3]?.trim();
        type = "a"; // arrow function
      }
    } else {
      // Class method or object method
      const methodMatch = fullMatch.match(
        /(?:(?:public|private|protected|static)\s+)?(?:async\s+)?(?:get\s+|set\s+)?(\w+)\s*\(([^)]*)\)(?:\s*:\s*([^{]+?))?/
      );
      if (methodMatch) {
        name = methodMatch[1];
        params = this.parseParameters(methodMatch[2]);
        returnType = methodMatch[3]?.trim();
        type = "m"; // method

        if (fullMatch.includes("get ")) type = "g"; // getter
        if (fullMatch.includes("set ")) type = "s"; // setter
        if (name === "constructor") type = "c"; // constructor
      }
    }

    // Determine if exported (basic heuristic)
    if (!exported) {
      // Check if function is at module level and not in a class/function
      const beforeMatch = this.content.substring(0, match.index);
      const braceCount =
        (beforeMatch.match(/\{/g) || []).length -
        (beforeMatch.match(/\}/g) || []).length;
      exported = braceCount === 0 && !name.startsWith("_");
    }

    if (name) {
      const funcInfo = new FunctionInfo({
        file: this.relativePath,
        line: lineNumber,
        type,
        exported,
        name,
        className:
          this.classStack.length > 0
            ? this.classStack[this.classStack.length - 1]
            : null,
        params,
        returns: returnType,
        decorators: [], // Would need more sophisticated parsing for decorators
        isAsync,
        isGenerator,
      });

      this.functions.push(funcInfo);
    }
  }

  parseParameters(paramString) {
    if (!paramString || paramString.trim() === "") {
      return [];
    }

    return paramString
      .split(",")
      .map((param) => {
        return param.trim().replace(/\s*=\s*[^,]*/, ""); // Remove default values
      })
      .filter((param) => param.length > 0);
  }

  getLineNumber(index) {
    const beforeMatch = this.content.substring(0, index);
    return beforeMatch.split("\n").length;
  }
}

function extractFunctions(directory) {
  const functions = [];

  function walkDirectory(dir) {
    const files = fs.readdirSync(dir);

    for (const file of files) {
      const filePath = path.join(dir, file);
      const stat = fs.statSync(filePath);

      if (stat.isDirectory()) {
        // Skip common non-source directories
        if (
          !file.startsWith(".") &&
          !["node_modules", "dist", "build", "coverage"].includes(file)
        ) {
          walkDirectory(filePath);
        }
      } else if (file.match(/\.(js|jsx|ts|tsx|mjs|cjs)$/)) {
        const relativePath = path.relative(directory, filePath);
        const extractor = new JSFunctionExtractor(filePath, relativePath);
        functions.push(...extractor.extract());
      }
    }
  }

  walkDirectory(directory);
  return functions;
}

function formatFunction(fn) {
  const exported = fn.exported ? "y" : "n";

  // Build signature
  let signature = `(${fn.params.join(",")}`;

  if (fn.returns) {
    signature += `)${fn.returns}`;
  } else {
    signature += ")";
  }

  // Add async/generator prefixes
  const prefixes = [];
  if (fn.isAsync) prefixes.push("async");
  if (fn.isGenerator) prefixes.push("generator");

  if (prefixes.length > 0) {
    signature = `${prefixes.join(" ")} ${signature}`;
  }

  // Add decorators if any
  const decoratorsStr = fn.decorators.length > 0 ? fn.decorators.join(",") : "";

  // Format based on whether it's a class method or function
  if (fn.className) {
    return `${fn.file}:${fn.line}:${fn.type}:${exported}:${fn.name}:${fn.className}:${signature}:${decoratorsStr}`;
  } else {
    return `${fn.file}:${fn.line}:${fn.type}:${exported}:${fn.name}::${signature}:${decoratorsStr}`;
  }
}

function main() {
  const args = process.argv.slice(2);
  let directory = ".";

  // Parse command line arguments
  for (let i = 0; i < args.length; i++) {
    if (args[i] === "--dir" && i + 1 < args.length) {
      directory = args[i + 1];
      i++; // Skip next argument
    }
  }

  try {
    const functions = extractFunctions(directory);

    // Sort by file, then by line number (same as other tools)
    functions.sort((a, b) => {
      if (a.file !== b.file) {
        return a.file.localeCompare(b.file);
      }
      return a.line - b.line;
    });

    functions.forEach((fn) => {
      console.log(formatFunction(fn));
    });
  } catch (error) {
    console.error(`Error: ${error.message}`);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

module.exports = { JSFunctionExtractor, extractFunctions, formatFunction };
