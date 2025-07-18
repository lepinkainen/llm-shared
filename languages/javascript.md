# JavaScript/TypeScript Guidelines

When looking for functions, use the `jsfuncs` tool to list all functions in a JavaScript/TypeScript project. It provides a compact format that is optimized for LLM context.

- Example usage:

  ```bash
  node jsfuncs.js --dir /path/to/project
  ```

## JavaScript/TypeScript Libraries and Tools

### Package Management

- Use "pnpm" for package management (faster than npm/yarn)
- Alternative: "npm" or "yarn" for simpler projects

### Development Tools

- **TypeScript**: Always prefer TypeScript over JavaScript for new projects
- **ESLint**: For linting (`@typescript-eslint/parser` for TS projects)
- **Prettier**: For code formatting
- **Vitest**: For testing (modern alternative to Jest)
- **tsx**: For running TypeScript directly (`npx tsx script.ts`)

### Framework Preferences

- **Node.js Backend**: Fastify (lightweight) or Express.js (established)
- **Frontend**: React with TypeScript, Next.js for full-stack
- **Build Tools**: Vite (modern) or esbuild (fast)

### Database Libraries

- **SQL**: Drizzle ORM (type-safe) or Prisma (feature-rich)
- **SQLite**: better-sqlite3
- **PostgreSQL**: pg with @types/pg

### Utility Libraries

- **Date handling**: date-fns (modular) or dayjs (lightweight)
- **Validation**: zod (TypeScript-first schema validation)
- **Environment**: dotenv for configuration
- **Logging**: pino (fast structured logging)

## General Guidelines for Javascript/TypeScript

- Always use TypeScript for new projects
- Configure strict TypeScript settings in tsconfig.json
- Use ES modules (import/export) over CommonJS
- Prefer async/await over promises and callbacks
- Use meaningful variable and function names
- Always build/compile before finishing a task