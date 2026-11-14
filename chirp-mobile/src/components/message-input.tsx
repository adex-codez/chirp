import { useEffect } from "react";
import { StyleSheet, TextInput, View } from "react-native";
import { useThemeColor } from "heroui-native/hooks";
import Animated, {
  Extrapolation,
  interpolate,
  useAnimatedStyle,
  useSharedValue,
  withTiming,
} from "react-native-reanimated";

import { Mic, Paperclip } from "@/components/icons";
import { dawn } from "@/theme/colors";
import { motion } from "@/theme/motion";
import { shadows } from "@/theme/shadows";
import { spacing } from "@/theme/spacing";

type MessageInputProps = { compact?: boolean };

// Matches the `h-11` circles; SLIDE is how far each one tucks behind the field.
const CIRCLE = 44;
const SLIDE = 12;

/* Every element is `h-11`, so the tallest one is the row's footprint. */
export const COMPOSER_HEIGHT = CIRCLE;

/*
 * Floating composer. The field and its two circles stand apart at one height,
 * so the thread scrolls behind three separate surfaces and clips under their
 * bottom edge. Both circles stay inert until the messaging layer lands, so
 * they do not swallow taps as Pressables would.
 *
 * While the thread is scrolled back, the circles tuck behind their slots, the
 * gap collapses, and the field takes the full width.
 */
export function MessageInput({ compact = false }: MessageInputProps) {
  const [foreground, accentForeground, placeholder] = useThemeColor([
    "foreground",
    "accent-foreground",
    "field-placeholder",
  ]);
  const progress = useSharedValue(compact ? 1 : 0);

  useEffect(() => {
    progress.value = withTiming(compact ? 1 : 0, { duration: motion.base });
  }, [compact, progress]);

  const attachSlotStyle = useAnimatedStyle(() => ({
    width: interpolate(progress.value, [0, 1], [CIRCLE, 0], Extrapolation.CLAMP),
    opacity: interpolate(progress.value, [0, 0.7], [1, 0], Extrapolation.CLAMP),
  }));

  const attachSlideStyle = useAnimatedStyle(() => ({
    transform: [{ translateX: interpolate(progress.value, [0, 1], [0, SLIDE], Extrapolation.CLAMP) }],
  }));

  const micSlotStyle = useAnimatedStyle(() => ({
    width: interpolate(progress.value, [0, 1], [CIRCLE, 0], Extrapolation.CLAMP),
    opacity: interpolate(progress.value, [0, 0.7], [1, 0], Extrapolation.CLAMP),
  }));

  const micSlideStyle = useAnimatedStyle(() => ({
    transform: [
      { translateX: interpolate(progress.value, [0, 1], [0, -SLIDE], Extrapolation.CLAMP) },
    ],
  }));

  const fieldStyle = useAnimatedStyle(() => ({
    marginHorizontal: interpolate(
      progress.value,
      [0, 1],
      [spacing.sm, 0],
      Extrapolation.CLAMP,
    ),
  }));

  const slotsInert = compact
    ? {
        pointerEvents: "none" as const,
        accessibilityElementsHidden: true,
        importantForAccessibility: "no-hide-descendants" as const,
      }
    : {
        pointerEvents: "auto" as const,
        accessibilityElementsHidden: false,
        importantForAccessibility: "auto" as const,
      };

  return (
    <View className="flex-row items-center">
      <Animated.View style={[styles.circleSlot, styles.attachSlot, attachSlotStyle]} {...slotsInert}>
        <Animated.View style={attachSlideStyle}>
          <View className="h-11 w-11 items-center justify-center rounded-full border border-separator bg-surface">
            <Paperclip color={foreground} />
          </View>
        </Animated.View>
      </Animated.View>

      <Animated.View style={[styles.field, fieldStyle]}>
        <View
          className="h-11 flex-row items-center rounded-full border border-separator bg-surface px-4"
          style={{ boxShadow: shadows.raised }}
        >
          <TextInput
            accessibilityLabel="Message"
            placeholder="Message"
            placeholderTextColor={placeholder}
            selectionColor={dawn.DEFAULT}
            className="h-9 flex-1 text-base text-foreground"
          />
        </View>
      </Animated.View>

      <Animated.View style={[styles.circleSlot, styles.micSlot, micSlotStyle]} {...slotsInert}>
        <Animated.View style={micSlideStyle}>
          <View className="h-11 w-11 items-center justify-center rounded-full bg-accent">
            <Mic color={accentForeground} />
          </View>
        </Animated.View>
      </Animated.View>
    </View>
  );
}

const styles = StyleSheet.create({
  circleSlot: {
    height: CIRCLE,
    overflow: "hidden",
    justifyContent: "center",
  },
  attachSlot: { alignItems: "flex-start" },
  micSlot: { alignItems: "flex-end" },
  field: { flex: 1 },
});
