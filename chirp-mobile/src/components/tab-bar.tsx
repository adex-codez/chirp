import type { ComponentType } from "react";
import { Pressable, View, useWindowDimensions } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import Svg, { Path } from "react-native-svg";
import { useThemeColor } from "heroui-native/hooks";
import { Typography } from "heroui-native/text";

import type { IconProps } from "@/components/icons";
import { Broadcast, Message, Phone, Sliders, Users } from "@/components/icons";
import { shadows } from "@/theme/shadows";
import { barLabel } from "@/theme/typography";

/* The bar's own geometry: a shape, not spacing tokens. */
const BAR = { plate: 70, well: 30, fillet: 8, circle: 52, icon: 22, lift: 26 } as const;

/* Offsets from the plate's top edge, shared by both slot shapes. */
const ICON_TOP = 15;
const LABEL_TOP = 44;
const TICK = { width: 12, height: 2, marginTop: 6 } as const;

/*
 * The raised slot starts above the plate so the whole floating circle is a
 * touch target. Its label still lands on LABEL_TOP, after the circle.
 */
const SLOT = {
  flat: {
    height: BAR.plate,
    marginTop: 0,
    paddingTop: ICON_TOP,
    labelGap: LABEL_TOP - ICON_TOP - BAR.icon,
  },
  raised: {
    height: BAR.plate + BAR.lift,
    marginTop: -BAR.lift,
    paddingTop: 0,
    labelGap: LABEL_TOP + BAR.lift - BAR.circle,
  },
} as const;

export type TabDestination = {
  route: string;
  label: string;
  Icon: ComponentType<IconProps>;
  raised?: boolean;
};

export const TAB_DESTINATIONS: readonly TabDestination[] = [
  { route: "status", label: "Status", Icon: Broadcast },
  { route: "community", label: "Community", Icon: Users },
  { route: "index", label: "Chat", Icon: Message, raised: true },
  { route: "call", label: "Call", Icon: Phone },
  { route: "settings", label: "Settings", Icon: Sliders },
];

const RAISED_INDEX = TAB_DESTINATIONS.findIndex((destination) => destination.raised === true);

const round = (value: number) => value.toFixed(3);

export function tabBarPath({
  width,
  height,
  notchX,
}: {
  width: number;
  height: number;
  notchX: number;
}): { fill: string; edge: string } {
  const top = 0.5;
  const dx = Math.sqrt(BAR.well ** 2 + 2 * BAR.well * BAR.fillet);
  const px = (BAR.well * dx) / (BAR.well + BAR.fillet);
  const py = (BAR.well * BAR.fillet) / (BAR.well + BAR.fillet);

  const contour =
    `M0 ${round(top)} H${round(notchX - dx)}` +
    ` A${round(BAR.fillet)} ${round(BAR.fillet)} 0 0 1 ${round(notchX - px)} ${round(top + py)}` +
    ` A${round(BAR.well)} ${round(BAR.well)} 0 0 0 ${round(notchX + px)} ${round(top + py)}` +
    ` A${round(BAR.fillet)} ${round(BAR.fillet)} 0 0 1 ${round(notchX + dx)} ${round(top)}` +
    ` H${round(width)}`;

  return { fill: `${contour} V${round(height)} H0 Z`, edge: contour };
}

export function useTabBarSpace(): number {
  const insets = useSafeAreaInsets();
  return BAR.plate + insets.bottom;
}

type TabBarProps = {
  state: { index: number; routes: { key: string; name: string }[] };
  navigation: { navigate: (name: string) => void };
  insets: { bottom: number };
};

export function TabBar({ state, navigation, insets }: TabBarProps) {
  const { width } = useWindowDimensions();
  const [surface, border, muted, accent, accentForeground] = useThemeColor([
    "surface",
    "border",
    "muted",
    "accent",
    "accent-foreground",
  ]);
  const height = BAR.plate + insets.bottom;
  const notchX = (width * (RAISED_INDEX + 0.5)) / TAB_DESTINATIONS.length;
  const { fill, edge } = tabBarPath({ width, height, notchX });

  return (
    <View style={{ position: "absolute", left: 0, right: 0, bottom: 0, height }}>
      <Svg
        width={width}
        height={height}
        pointerEvents="none"
        style={{ position: "absolute", left: 0, top: 0 }}
      >
        <Path d={fill} fill={surface} />
        <Path d={edge} fill="none" stroke={border} strokeWidth={1} />
      </Svg>

      <View style={{ height: BAR.plate, flexDirection: "row" }}>
        {TAB_DESTINATIONS.map((destination) => {
          const { Icon } = destination;
          const selected = state.routes[state.index]?.name === destination.route;
          const layout = destination.raised ? SLOT.raised : SLOT.flat;

          return (
            <Pressable
              key={destination.route}
              accessibilityRole="tab"
              accessibilityState={{ selected }}
              accessibilityLabel={destination.label}
              onPress={() => navigation.navigate(destination.route)}
              style={({ pressed }) => [
                { flex: 1, alignItems: "center" },
                layout,
                pressed ? { opacity: 0.6 } : null,
              ]}
            >
              {destination.raised ? (
                <View
                  style={{
                    width: BAR.circle,
                    height: BAR.circle,
                    borderRadius: BAR.circle / 2,
                    borderWidth: 1,
                    alignItems: "center",
                    justifyContent: "center",
                    backgroundColor: selected ? accent : surface,
                    borderColor: selected ? accent : border,
                    boxShadow: shadows.card,
                  }}
                >
                  <Icon color={selected ? accentForeground : muted} />
                </View>
              ) : (
                <Icon color={selected ? accent : muted} />
              )}

              <Typography
                style={[barLabel, { marginTop: layout.labelGap, color: selected ? accent : muted }]}
                type="body-xs"
                weight={selected ? "bold" : "normal"}
              >
                {destination.label.toUpperCase()}
              </Typography>

              {selected ? (
                <View
                  style={{
                    width: TICK.width,
                    height: TICK.height,
                    marginTop: TICK.marginTop,
                    borderRadius: TICK.height / 2,
                    backgroundColor: accent,
                  }}
                />
              ) : null}
            </Pressable>
          );
        })}
      </View>
    </View>
  );
}
