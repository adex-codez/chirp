import { View } from "react-native";
import Svg, { Circle } from "react-native-svg";
import { Typography } from "heroui-native/text";
import { dawn, lichen } from "@/theme/colors";
import { fontStack } from "@/theme/typography";

/*
 * BandChip — a Username rendered as an aluminum leg band.
 * Rivet dots + mono type make verified identity tangible.
 * `verified` tints the rivets lichen green.
 */

export function BandChip({ username, verified = false }: { username: string; verified?: boolean }) {
  const rivet = verified ? lichen : dawn.DEFAULT;
  return (
    <View className="flex-row items-center gap-2 self-start rounded-full border border-separator bg-surface px-3 py-1.5">
      <Svg width={8} height={8} viewBox="0 0 8 8">
        <Circle cx={4} cy={4} r={3.2} fill={rivet} />
      </Svg>
      <Typography style={{ fontFamily: fontStack.band }} type="body-sm" weight="bold">
        @{username}
      </Typography>
      <Svg width={8} height={8} viewBox="0 0 8 8">
        <Circle cx={4} cy={4} r={3.2} fill={rivet} />
      </Svg>
    </View>
  );
}
