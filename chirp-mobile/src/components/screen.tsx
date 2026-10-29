import type { ReactNode } from "react";
import { KeyboardAvoidingView, Platform, ScrollView, View } from "react-native";
import { StatusBar } from "expo-status-bar";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { SafeAreaView } from "@/components/safe-area-view";
import { spacing } from "@/theme/spacing";

/*
 * `edges` defaults to top only: the content pads itself by the bottom
 * inset, so claiming that edge in the safe area too would double the gap.
 */

type ScreenProps = {
  children: ReactNode;
  className?: string;
  contentClassName?: string;
  scroll?: boolean;
  edges?: ("top" | "bottom")[];
  /** Space reserved for an overlay at the bottom: pass the overlay's full height. */
  bottomSpace?: number;
};

export function Screen({
  children,
  className,
  contentClassName,
  scroll = true,
  edges = ["top"],
  bottomSpace,
}: ScreenProps) {
  const insets = useSafeAreaInsets();
  const paddingBottom = Math.max(insets.bottom, bottomSpace ?? 0) + spacing.lg;

  return (
    <SafeAreaView edges={edges} className={`flex-1 bg-background ${className ?? ""}`}>
      <StatusBar style="dark" />
      <KeyboardAvoidingView
        className="flex-1"
        behavior={Platform.OS === "ios" ? "padding" : undefined}
      >
        {scroll ? (
          <ScrollView
            className="flex-1"
            contentContainerClassName={`gap-6 px-6 pt-4 ${contentClassName ?? ""}`}
            contentContainerStyle={{ paddingBottom }}
            keyboardShouldPersistTaps="handled"
            keyboardDismissMode="on-drag"
          >
            {children}
          </ScrollView>
        ) : (
          <View
            className={`flex-1 gap-6 px-6 pt-4 ${contentClassName ?? ""}`}
            style={{ paddingBottom }}
          >
            {children}
          </View>
        )}
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}
