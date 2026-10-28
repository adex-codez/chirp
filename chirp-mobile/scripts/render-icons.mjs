import { execFileSync } from "node:child_process";
import { existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

/*
 * Rasterises the icon PNGs from the SVG sources in assets/images.
 * Uses headless Edge, so there is no image dependency to install.
 * Run with `npm run icons` after editing chirp-logo.svg or chirp-mark.svg.
 */

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const images = path.join(root, "assets", "images");

const EDGE = [
  "C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe",
  "C:/Program Files/Microsoft/Edge/Application/msedge.exe",
].find((candidate) => existsSync(candidate));

if (!EDGE) {
  throw new Error("Microsoft Edge not found; edit EDGE candidates in scripts/render-icons.mjs");
}

const lockup = readFileSync(path.join(images, "chirp-logo.svg"), "utf8");
const mark = readFileSync(path.join(images, "chirp-mark.svg"), "utf8");
const monoMark = mark.replaceAll("#FFC247", "#FFFFFF").replaceAll("#94B0C8", "#FFFFFF");

const dataUri = (svg) => `data:image/svg+xml;base64,${Buffer.from(svg, "utf8").toString("base64")}`;

const page = (svg, size, art) => `<!doctype html>
<html><head><meta charset="utf-8"><style>
  html, body { margin: 0; background: transparent; }
  body { width: ${size}px; height: ${size}px; display: grid; place-items: center; overflow: hidden; }
  img { display: block; width: ${art}px; height: ${art}px; }
</style></head>
<body><img src="${dataUri(svg)}" alt=""></body></html>`;

const targets = [
  { file: "icon.png", size: 1024, svg: lockup, art: 1024 },
  { file: "splash-icon.png", size: 512, svg: mark, art: 400 },
  { file: "favicon.png", size: 64, svg: lockup, art: 64 },
  { file: "android-icon-foreground.png", size: 1024, svg: mark, art: 803 },
  { file: "android-icon-monochrome.png", size: 1024, svg: monoMark, art: 803 },
];

const work = mkdtempSync(path.join(tmpdir(), "chirp-icons-"));

try {
  for (const target of targets) {
    const html = path.join(work, `${target.file}.html`);
    writeFileSync(html, page(target.svg, target.size, target.art));

    execFileSync(
      EDGE,
      [
        "--headless=new",
        "--disable-gpu",
        "--hide-scrollbars",
        "--force-device-scale-factor=1",
        `--window-size=${target.size},${target.size}`,
        "--default-background-color=00000000",
        `--screenshot=${path.join(images, target.file)}`,
        `file:///${html.replaceAll("\\", "/")}`,
      ],
      { stdio: "ignore" },
    );

    console.log(`wrote ${target.file} (${target.size}x${target.size})`);
  }
} finally {
  rmSync(work, { recursive: true, force: true });
}
