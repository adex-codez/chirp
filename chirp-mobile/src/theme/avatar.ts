import { accent } from "./colors";

export type AvatarInk = {
  readonly background: string;
  readonly foreground: string;
};

const AVATAR_SATURATION = 38;
const AVATAR_LIGHTNESS = 32;

export function avatarInk(seed: string): AvatarInk {
  let hash = 0x811c9dc5;
  for (let index = 0; index < seed.length; index += 1) {
    hash ^= seed.charCodeAt(index);
    hash = Math.imul(hash, 0x01000193);
  }
  const hue = (hash >>> 0) % 360;
  return {
    background: `hsl(${hue}, ${AVATAR_SATURATION}%, ${AVATAR_LIGHTNESS}%)`,
    foreground: accent.light.foreground,
  };
}
