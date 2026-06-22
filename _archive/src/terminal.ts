import { spawn } from "node:child_process";
import readline from "node:readline";

export const ansi = {
  clear: "\x1b[2J\x1b[H",
  hideCursor: "\x1b[?25l",
  showCursor: "\x1b[?25h",
  reset: "\x1b[0m",
  bold: "\x1b[1m",
  dim: "\x1b[2m",
  inverse: "\x1b[7m",
  red: "\x1b[31m",
  green: "\x1b[32m",
  yellow: "\x1b[33m",
  blue: "\x1b[34m",
  cyan: "\x1b[36m",
  gray: "\x1b[90m",
};

export function enterAltScreen(): void {
  process.stdout.write("\x1b[?1049h" + ansi.hideCursor);
}

export function leaveAltScreen(): void {
  process.stdout.write(ansi.showCursor + "\x1b[?1049l");
}

export function terminalSize(): { columns: number; rows: number } {
  return {
    columns: process.stdout.columns || 100,
    rows: process.stdout.rows || 32,
  };
}

export function truncate(value: string | undefined, width: number): string {
  const text = value ?? "";
  if (width <= 0) return "";
  if (text.length <= width) return text.padEnd(width);
  if (width === 1) return "…";
  return `${text.slice(0, width - 1)}…`;
}

export function stripAnsi(value: string): string {
  return value.replace(/\x1b\[[0-9;?]*[A-Za-z]/g, "");
}

export function visibleLength(value: string): number {
  return stripAnsi(value).length;
}

export function openUrl(url: string): void {
  const command = process.platform === "darwin" ? "open" : process.platform === "win32" ? "cmd" : "xdg-open";
  const args = process.platform === "win32" ? ["/c", "start", "", url] : [url];
  const child = spawn(command, args, { detached: true, stdio: "ignore" });
  child.unref();
}

export function setupKeypress(onKey: (key: string) => void): () => void {
  readline.emitKeypressEvents(process.stdin);
  const wasRaw = process.stdin.isRaw;
  if (process.stdin.isTTY) process.stdin.setRawMode(true);

  const handler = (_str: string, key: readline.Key) => {
    if (key.ctrl && key.name === "c") {
      onKey("ctrl+c");
      return;
    }
    if (key.name) onKey(key.shift ? `shift+${key.name}` : key.name);
  };

  process.stdin.on("keypress", handler);
  return () => {
    process.stdin.off("keypress", handler);
    if (process.stdin.isTTY) process.stdin.setRawMode(wasRaw);
  };
}
