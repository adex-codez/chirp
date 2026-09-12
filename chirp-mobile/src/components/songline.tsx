import Svg, { Rect } from "react-native-svg";

import { dawn } from "@/theme/colors";

/*
 * Songline — a chirp sonogram. The single signature element: gold bars
 * shaped like a spectrogram slice, used as divider, step meter, and
 * verified-seal echo. Bars are data, not decoration: `levels` encodes
 * the shape, `active` fills progress through the auth sequence.
 */

type SonglineProps = {
  levels?: number[];
  active?: number;
  height?: number;
};

const DEFAULT_LEVELS = [8, 16, 26, 14, 32, 22, 38, 18, 28, 12, 24, 34, 16, 26, 10, 20, 30, 14, 22, 8];

export function Songline({ levels = DEFAULT_LEVELS, active, height = 40 }: SonglineProps) {
  const barWidth = 6;
  const gap = 5;
  const width = levels.length * (barWidth + gap);

  return (
    <Svg
      width={width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      accessibilityRole="image"
      accessibilityLabel="Chirp sonogram"
    >
      {levels.map((level, i) => {
        const h = Math.max(4, Math.min(height, level));
        const y = (height - h) / 2;
        const isOn = active === undefined || i < active;
        return (
          <Rect
            key={i}
            x={i * (barWidth + gap)}
            y={y}
            width={barWidth}
            height={h}
            rx={3}
            fill={isOn ? dawn.DEFAULT : "#FFFFFF"}
            opacity={isOn ? 1 : 0.22}
          />
        );
      })}
    </Svg>
  );
}
