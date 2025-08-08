# TypeScript/JavaScript Guidelines

**Default Runtime**: Always use TypeScript as the default language for all new projects. JavaScript should only be used for legacy maintenance or specific constraints.

When looking for functions, use the `jsfuncs` tool to list all functions in a JavaScript/TypeScript project. It provides a compact format that is optimized for LLM context.

- Example usage:

  ```bash
  node jsfuncs.js --dir /path/to/project
  ```

## Runtime Preferences

### Deno (Preferred)

- **Default choice** for new TypeScript projects
- No package.json required - uses import maps and npm: specifiers
- Built-in TypeScript support (no compilation step)
- Web-standard APIs (fetch, crypto, etc.)
- Built-in testing, formatting, and linting
- Secure by default (explicit permissions)
- Example: `deno run -A server.ts`

### Node.js (Legacy/Compatibility)

- Use only when Deno compatibility is not feasible
- Requires separate TypeScript compilation step
- Package management via pnpm (preferred) or npm

## TypeScript Libraries and Tools

### Package Management (Node.js projects)

- **pnpm**: Preferred for Node.js projects (faster than npm/yarn)
- **npm**: Fallback for simpler projects

### Development Tools

- **Deno**: Built-in TypeScript, testing, formatting, linting
- **ESLint**: For Node.js projects (`@typescript-eslint/parser`)
- **Prettier**: For Node.js projects (Deno has built-in `deno fmt`)
- **Vitest**: For Node.js testing (modern alternative to Jest)
- **tsx**: For running TypeScript in Node.js (`npx tsx script.ts`)

### Framework Preferences

- **Deno Backend**: Use standard library HTTP server or Oak framework
- **Node.js Backend**: Fastify (lightweight) or Express.js (established)
- **Frontend**: React with TypeScript, Next.js for full-stack
- **Build Tools**: Vite (modern) or esbuild (fast)

### Database Libraries

#### Deno Databases

- **SQLite**: Use `npm:better-sqlite3` or Deno's built-in sqlite
- **PostgreSQL**: Use `npm:postgres` or `npm:pg` with types
- **ORM**: Consider `npm:drizzle-orm` for type-safe queries

#### Node.js Databases

- **SQL**: Drizzle ORM (type-safe) or Prisma (feature-rich)
- **SQLite**: better-sqlite3
- **PostgreSQL**: pg with @types/pg

### Utility Libraries

#### Deno Utilities

- **Date handling**: Use built-in Date APIs or `npm:date-fns`
- **Validation**: `npm:zod` (TypeScript-first schema validation)
- **Environment**: Use `Deno.env.get()` (built-in)
- **Logging**: Use `console` or `npm:pino`
- **HTTP client**: Use built-in `fetch()`
- **Crypto**: Use built-in `crypto` Web API

#### Node.js Utilities

- **Date handling**: date-fns (modular) or dayjs (lightweight)
- **Validation**: zod (TypeScript-first schema validation)
- **Environment**: dotenv for configuration
- **Logging**: pino (fast structured logging)

## General TypeScript Guidelines

- **Always use TypeScript** as the default - no exceptions for new projects
- **Deno first**: Prefer Deno runtime for new projects unless Node.js is required
- **Strict mode**: Configure strict TypeScript settings (`strict: true`)
- **ES modules**: Always use ES modules (import/export) over CommonJS
- **Modern syntax**: Prefer async/await over promises and callbacks
- **Type safety**: Leverage TypeScript's type system fully - avoid `any`
- **Web standards**: Prefer web-standard APIs (fetch, crypto, etc.) when available
- **Zero config**: Leverage Deno's zero-config approach when possible
