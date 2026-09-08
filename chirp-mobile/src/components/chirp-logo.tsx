import Svg, {
  Circle,
  Ellipse,
  Path,
  Polygon,
} from "react-native-svg";

import { brand } from "@/theme/colors";

/*
 * Chirp bird mark — original artwork, mirrors
 * `assets/images/chirp-mark.svg` (512 space).
 * Use for in-app surfaces (loading, empty states, headers).
 * The OS app icon / splash PNGs are raster exports of the same art.
 */

type ChirpLogoProps = {
  size?: number;
  /** Single-color silhouette for themed/adaptive contexts. */
  monochrome?: boolean;
};

const NAVY = brand[800];
const MIST = brand[300];
const BELLY = brand[100];
const GOLD = "#FFC247";
const GOLD_DARK = "#E89B2E";

export function ChirpLogo({ size = 64, monochrome = false }: ChirpLogoProps) {
  const body = "#FFFFFF";
  const wing = monochrome ? "#FFFFFF" : MIST;
  const beakUp = monochrome ? "#FFFFFF" : GOLD;
  const beakLo = monochrome ? "#FFFFFF" : GOLD_DARK;

  return (
    <Svg
      width={size}
      height={size}
      viewBox="0 0 512 512"
      accessibilityRole="image"
      accessibilityLabel="Chirp bird logo"
    >
      <Path
        d="M150 235 L78 198 L96 252 L64 272 L98 300 L80 352 L150 320 Z"
        fill={body}
        stroke={body}
        strokeWidth={18}
        strokeLinejoin="round"
      />
      <Ellipse cx={245} cy={285} rx={118} ry={128} fill={body} />
      {monochrome ? null : (
        <Ellipse cx={250} cy={360} rx={70} ry={34} fill={BELLY} opacity={0.9} />
      )}
      <Path
        d="M178 262 C226 248 282 272 300 332 C254 328 200 308 178 262 Z"
        fill={wing}
      />
      <Polygon
        points="348,252 418,232 354,282"
        fill={beakUp}
        stroke={beakUp}
        strokeWidth={10}
        strokeLinejoin="round"
      />
      <Polygon
        points="352,290 404,300 348,308"
        fill={beakLo}
        stroke={beakLo}
        strokeWidth={10}
        strokeLinejoin="round"
      />
      {monochrome ? null : (
        <>
          <Circle cx={312} cy={228} r={17} fill={NAVY} />
          <Circle cx={318} cy={222} r={5.5} fill="#FFFFFF" />
        </>
      )}
      <Path
        d="M447.6 228.2 A48 48 0 0 1 447.6 303.8"
        fill="none"
        stroke={body}
        strokeWidth={15}
        strokeLinecap="round"
      />
      <Path
        d="M472.2 196.7 A88 88 0 0 1 472.2 335.3"
        fill="none"
        stroke={body}
        strokeWidth={15}
        strokeLinecap="round"
        opacity={monochrome ? 1 : 0.55}
      />
    </Svg>
  );
}
