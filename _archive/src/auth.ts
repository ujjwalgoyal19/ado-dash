import { execFile } from "node:child_process";
import { promisify } from "node:util";
import type { AuthConfig } from "./types.js";

const execFileAsync = promisify(execFile);
const AZURE_DEVOPS_RESOURCE_ID = "499b84ac-1321-427f-aa17-267ca6975798";

export interface AuthProvider {
  getAuthorizationHeader(): Promise<string>;
}

export function createAuthProvider(config: AuthConfig): AuthProvider {
  if (config.type === "pat") return new PatAuthProvider(config.env ?? "AZURE_DEVOPS_PAT");
  return new AzureCliAuthProvider(config.tenant);
}

class PatAuthProvider implements AuthProvider {
  constructor(private readonly envName: string) {}

  async getAuthorizationHeader(): Promise<string> {
    const pat = process.env[this.envName];
    if (!pat) throw new Error(`Missing Azure DevOps PAT. Set ${this.envName}=<token>.`);
    return `Basic ${Buffer.from(`:${pat}`, "utf8").toString("base64")}`;
  }
}

interface AzureCliTokenResponse {
  accessToken?: string;
  expiresOn?: string;
  expires_on?: number;
}

class AzureCliAuthProvider implements AuthProvider {
  private cached?: { token: string; expiresAt: number };

  constructor(private readonly tenant?: string) {}

  async getAuthorizationHeader(): Promise<string> {
    const token = await this.getToken();
    return `Bearer ${token}`;
  }

  private async getToken(): Promise<string> {
    if (this.cached && Date.now() < this.cached.expiresAt - 5 * 60 * 1000) {
      return this.cached.token;
    }

    const args = [
      "account",
      "get-access-token",
      "--resource",
      AZURE_DEVOPS_RESOURCE_ID,
      "--output",
      "json",
    ];
    if (this.tenant) args.push("--tenant", this.tenant);

    try {
      const { stdout } = await execFileAsync("az", args, { maxBuffer: 1024 * 1024 });
      const parsed = JSON.parse(stdout) as AzureCliTokenResponse;
      if (!parsed.accessToken) throw new Error("Azure CLI response did not include accessToken");

      this.cached = {
        token: parsed.accessToken,
        expiresAt: parseAzureCliExpiry(parsed),
      };
      return this.cached.token;
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      throw new Error(
        `Azure CLI auth failed. Run \`az login\` in a browser first, then retry. Details: ${message}`,
      );
    }
  }
}

function parseAzureCliExpiry(response: AzureCliTokenResponse): number {
  if (typeof response.expires_on === "number" && Number.isFinite(response.expires_on)) {
    return response.expires_on * 1000;
  }

  if (response.expiresOn) {
    const parsed = Date.parse(response.expiresOn);
    if (Number.isFinite(parsed)) return parsed;
  }

  return Date.now() + 55 * 60 * 1000;
}
