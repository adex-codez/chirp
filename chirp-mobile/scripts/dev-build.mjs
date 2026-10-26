#!/usr/bin/env node
import { spawnSync } from "node:child_process";

const PLATFORMS = ["android", "ios", "all"];
const PROFILE = "development";

const HELP = [
  "Usage: bun dev-build [options] [-- <eas flags>]",
  "",
  "Builds the EAS development client. Install the artifact on a device,",
  "then run `bun start` to load the JS bundle over the tunnel.",
  "",
  "Options:",
  "  -p, --platform <android|ios|all>  Platform to build. Default: all",
  "      --dry-run                     Print the EAS command and exit",
  "  -h, --help                        Show this message",
  "",
  "Any other flag reaches `eas build`, for example --local or --no-wait.",
  "iOS builds run on EAS, so they work from Windows and Linux too.",
].join("\n");

function parseArgs(argv) {
  const passthrough = [];
  let platform = "all";
  let dryRun = false;
  let help = false;

  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "-h" || arg === "--help") {
      help = true;
    } else if (arg === "--dry-run") {
      dryRun = true;
    } else if (arg === "-p" || arg === "--platform") {
      platform = argv[i + 1] ?? "";
      i += 1;
    } else if (arg.startsWith("--platform=")) {
      platform = arg.slice("--platform=".length);
    } else if (arg === "--") {
      passthrough.push(...argv.slice(i + 1));
      break;
    } else {
      passthrough.push(arg);
    }
  }

  return { platform, dryRun, help, passthrough };
}

const { platform, dryRun, help, passthrough } = parseArgs(process.argv.slice(2));

if (help) {
  console.log(HELP);
  process.exit(0);
}

if (!PLATFORMS.includes(platform)) {
  console.error(`Unknown platform "${platform}". Use one of ${PLATFORMS.join(", ")}.`);
  process.exit(1);
}

const buildArgs = [
  "build",
  "--profile",
  PROFILE,
  "--platform",
  platform,
  ...passthrough,
];

if (dryRun) {
  console.log(`eas ${buildArgs.join(" ")}`);
  process.exit(0);
}

function shell(command, options = {}) {
  return spawnSync(command, { stdio: "inherit", shell: true, ...options });
}

const hasEas = spawnSync("eas", ["--version"], {
  stdio: "ignore",
  shell: true,
}).status === 0;
const eas = hasEas ? "eas" : "npx eas-cli";

if (shell(`${eas} whoami`, { stdio: "ignore" }).status !== 0) {
  console.error(`EAS is not signed in. Run \`${eas} login\` first.`);
  process.exit(1);
}

const build = shell(`${eas} ${buildArgs.join(" ")}`);
process.exit(build.status ?? 1);
