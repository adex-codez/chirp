import { View } from "react-native";
import Svg, { Circle } from "react-native-svg";
import { useThemeColor } from "heroui-native/hooks";
import { Typography } from "heroui-native/text";

import { Check } from "@/components/icons";
import { dawn } from "@/theme/colors";
import { bandLabel } from "@/theme/typography";

/*
 * BandChip — a Username rendered as an aluminum leg band.
 * Rivet dots + mono type make verified identity tangible.
 * `verified` tints the rivets with the theme's success color.
 */

export function BandChip({ username, verified = false }: { username: string; verified?: boolean }) {
  const success = useThemeColor("success");
  const rivet = verified ? success : dawn.DEFAULT;

  return (
    <View
      className="flex-row items-center gap-2 self-start rounded-full border border-separator bg-surface px-3 py-1.5"
      accessibilityLabel={`@${username}${verified ? ", verified" : ""}`}
    >
      <Svg width={8} height={8} viewBox="0 0 8 8">
        <Circle cx={4} cy={4} r={3.2} fill={rivet} />
      </Svg>
      <Typography style={bandLabel} type="body-sm" weight="bold">
        @{username}
      </Typography>
      {verified ? <Check color={success} /> : null}
    </View>
  );
}
