import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { parse } from "yaml";
import type { AppConfig, AuthConfig, SectionConfig } from "./types.js";

const DEFAULT_CONFIG_NAMES = [".ado-dash.yml", ".ado-dash.yaml", "ado-dash.yml", "ado-dash.yaml"];

export function findConfigPath(explicitPath?: string): string {
  if (explicitPath) return explicitPath;
  if (process.env.ADO_DASH_CONFIG) return process.env.ADO_DASH_CONFIG;

  for (const name of DEFAULT_CONFIG_NAMES) {
    const candidate = join(process.cwd(), name);
    if (existsSync(candidate)) return candidate;
  }

  const xdg = process.env.XDG_CONFIG_HOME;
  if (xdg) {
    const candidate = join(xdg, "ado-dash", "config.yml");
    if (existsSync(candidate)) return candidate;
  }

  const home = process.env.HOME || process.env.USERPROFILE;
  if (home) {
    const candidate = join(home, ".config", "ado-dash", "config.yml");
    if (existsSync(candidate)) return candidate;
  }

  throw new Error(
    "No config found. Create .ado-dash.yml or pass --config /path/to/config.yml.",
  );
}

export function loadConfig(path: string): AppConfig {
  const raw = readFileSync(path, "utf8");
  const parsed = parse(raw) as unknown;
  return validateConfig(parsed, path);
}

function validateConfig(value: unknown, path: string): AppConfig {
  if (!isRecord(value)) throw new Error(`${path}: config must be a YAML object`);

  const organizationUrl = readString(value, "organizationUrl", path).replace(/\/+$/, "");
  const project = readString(value, "project", path);
  const patEnv = optionalString(value, "patEnv");
  const auth = validateAuth(value.auth, patEnv, path);
  const defaults = isRecord(value.defaults) ? {
    refreshSeconds: optionalNumber(value.defaults, "refreshSeconds"),
    limit: optionalNumber(value.defaults, "limit"),
  } : undefined;

  if (!Array.isArray(value.sections) || value.sections.length === 0) {
    throw new Error(`${path}: sections must be a non-empty list`);
  }

  const sections = value.sections.map((section, index) => validateSection(section, `${path}: sections[${index}]`));

  return { organizationUrl, project, auth, patEnv, defaults, sections };
}

function validateAuth(value: unknown, patEnv: string | undefined, path: string): AuthConfig {
  if (value === undefined || value === null) {
    return { type: "pat", env: patEnv ?? "AZURE_DEVOPS_PAT" };
  }

  if (typeof value === "string") {
    const type = oneOf(value, ["pat", "azure-cli"], `${path}.auth`)!;
    return type === "pat" ? { type, env: patEnv ?? "AZURE_DEVOPS_PAT" } : { type };
  }

  if (!isRecord(value)) throw new Error(`${path}.auth must be an object or string`);
  const type = oneOf(readString(value, "type", `${path}.auth`), ["pat", "azure-cli"], `${path}.auth.type`)!;

  if (type === "pat") {
    return { type, env: optionalString(value, "env") ?? patEnv ?? "AZURE_DEVOPS_PAT" };
  }

  return { type, tenant: optionalString(value, "tenant") };
}

function validateSection(value: unknown, label: string): SectionConfig {
  if (!isRecord(value)) throw new Error(`${label} must be an object`);
  const title = readString(value, "title", label);
  const type = readString(value, "type", label);
  const limit = optionalNumber(value, "limit");

  if (type === "pullRequests") {
    return {
      title,
      type,
      limit,
      status: oneOf(optionalString(value, "status"), ["active", "abandoned", "completed", "all"], `${label}.status`),
      role: oneOf(optionalString(value, "role"), ["author", "reviewer", "all"], `${label}.role`),
      repositoryId: optionalString(value, "repositoryId"),
      targetBranch: optionalString(value, "targetBranch"),
      sourceBranch: optionalString(value, "sourceBranch"),
    };
  }

  if (type === "workItems") {
    const wiql = readString(value, "wiql", label);
    const fields = Array.isArray(value.fields) ? value.fields.map((field, index) => {
      if (typeof field !== "string") throw new Error(`${label}.fields[${index}] must be a string`);
      return field;
    }) : undefined;
    return { title, type, limit, wiql, fields };
  }

  throw new Error(`${label}.type must be pullRequests or workItems`);
}

function readString(record: Record<string, unknown>, key: string, label: string): string {
  const value = record[key];
  if (typeof value !== "string" || value.trim() === "") throw new Error(`${label}.${key} must be a non-empty string`);
  return value;
}

function optionalString(record: Record<string, unknown>, key: string): string | undefined {
  const value = record[key];
  if (value === undefined || value === null) return undefined;
  if (typeof value !== "string") throw new Error(`${key} must be a string`);
  return value;
}

function optionalNumber(record: Record<string, unknown>, key: string): number | undefined {
  const value = record[key];
  if (value === undefined || value === null) return undefined;
  if (typeof value !== "number" || !Number.isFinite(value)) throw new Error(`${key} must be a number`);
  return value;
}

function oneOf<const T extends string>(value: string | undefined, options: readonly T[], label: string): T | undefined {
  if (value === undefined) return undefined;
  if ((options as readonly string[]).includes(value)) return value as T;
  throw new Error(`${label} must be one of: ${options.join(", ")}`);
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
