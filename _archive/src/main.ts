#!/usr/bin/env bun
import { createAuthProvider } from "./auth.js";
import { AzureDevOpsClient } from "./azureDevOps.js";
import { findConfigPath, loadConfig } from "./config.js";

interface CliArgs {
  configPath?: string;
  help: boolean;
  version: boolean;
  doctor: boolean;
}

async function main(): Promise<void> {
  const args = parseArgs(process.argv.slice(2));

  if (args.help) {
    printHelp();
    return;
  }

  if (args.version) {
    console.log("ado-dash 0.1.0");
    return;
  }

  const configPath = findConfigPath(args.configPath);
  const config = loadConfig(configPath);
  const authProvider = createAuthProvider(config.auth);

  const client = new AzureDevOpsClient(config, authProvider);

  if (args.doctor) {
    const result = await client.checkConnection();
    console.log(`Auth OK as: ${result.user.displayName}`);
    console.log(`Project OK: ${result.projectName ?? config.project}`);
    return;
  }

  const { OpenTuiDashboard } = await import("./openTuiDashboard.js");
  const dashboard = new OpenTuiDashboard(config, client);

  const cleanup = () => dashboard.stop();

  process.once("exit", cleanup);
  process.once("SIGINT", () => {
    cleanup();
    process.exit(130);
  });
  process.once("SIGTERM", () => {
    cleanup();
    process.exit(143);
  });

  await dashboard.run();
}

function parseArgs(argv: string[]): CliArgs {
  const result: CliArgs = { help: false, version: false, doctor: false };

  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];
    if (arg === "--help" || arg === "-h") result.help = true;
    else if (arg === "--version" || arg === "-v") result.version = true;
    else if (arg === "--doctor") result.doctor = true;
    else if (arg === "--config" || arg === "-c") {
      const value = argv[++i];
      if (!value) throw new Error(`${arg} requires a path`);
      result.configPath = value;
    } else if (arg?.startsWith("--config=")) {
      result.configPath = arg.slice("--config=".length);
    } else {
      throw new Error(`Unknown argument: ${arg}`);
    }
  }

  return result;
}

function printHelp(): void {
  console.log(`ado-dash

A terminal dashboard for Azure DevOps pull requests and work items.

Usage:
  ado-dash [--config .ado-dash.yml]
  ado-dash --doctor [--config .ado-dash.yml]

Auth:
  auth.type: azure-cli  Use browser login from \`az login\`.
  auth.type: pat        Use a PAT from auth.env / AZURE_DEVOPS_PAT.

Keys:
  tab/h/l switch sections
  j/k or arrows move
  r refresh current section
  R refresh all sections
  enter/o open selected item in browser
  v open selected PR in local review mode
  q quit
`);
}

main().catch((error) => {
  console.error(error instanceof Error ? error.message : error);
  process.exit(1);
});
