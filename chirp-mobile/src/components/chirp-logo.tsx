import { useId } from "react";
import Svg, { Circle, Defs, Mask, Path, Rect } from "react-native-svg";

import { dawn } from "@/theme/colors";

/*
 * Mirrors `assets/images/chirp-mark.svg` (512 space); the OS icon and
 * splash PNGs are raster exports of the same art, rendered by
 * `scripts/render-icons.mjs`.
 */

type ChirpLogoProps = {
  size?: number;
  /** Single-color silhouette for themed/adaptive contexts. */
  monochrome?: boolean;
};

const TRACE = "M 130 268 H 178 L 212 168 L 248 352 L 284 224 L 314 268 H 382";

export function ChirpLogo({ size = 64, monochrome = false }: ChirpLogoProps) {
  const maskId = `chirp-seal-${useId().replace(/[^a-zA-Z0-9-]/g, "")}`;

  return (
    <Svg
      width={size}
      height={size}
      viewBox="0 0 512 512"
      accessibilityRole="image"
      accessibilityLabel="Chirp logo"
    >
      <Defs>
        <Mask id={maskId} maskUnits="userSpaceOnUse" x={0} y={0} width={512} height={512}>
          <Rect width={512} height={512} fill="#000000" />
          <Circle cx={256} cy={256} r={196} fill="#FFFFFF" />
          <Path
            d={TRACE}
            fill="none"
            stroke="#000000"
            strokeWidth={46}
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </Mask>
      </Defs>
      <Rect
        width={512}
        height={512}
        fill={monochrome ? "#FFFFFF" : dawn.DEFAULT}
        mask={`url(#${maskId})`}
      />
    </Svg>
  );
}
