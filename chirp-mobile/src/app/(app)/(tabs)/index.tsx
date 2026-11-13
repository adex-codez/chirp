import { Pressable } from "react-native";
import { router } from "expo-router";
import { useThemeColor } from "heroui-native/hooks";
import { Typography } from "heroui-native/text";

import { Plus } from "@/components/icons";
import { Masthead } from "@/components/masthead";
import { Screen } from "@/components/screen";
import { useTabBarSpace } from "@/components/tab-bar";
import { spacing } from "@/theme/spacing";

export default function Chat() {
  const barSpace = useTabBarSpace();
  const accentForeground = useThemeColor("accent-foreground");

  return (
    <Screen
      bottomSpace={barSpace}
      overlay={
        <Pressable
          accessibilityRole="button"
          accessibilityLabel="Start a conversation"
          onPress={() => router.push("/contacts")}
          className="absolute right-6 h-14 w-14 items-center justify-center rounded-full bg-accent shadow-overlay"
          style={({ pressed }) => [
            { bottom: barSpace + spacing.md, borderCurve: "continuous" },
            pressed ? { opacity: 0.75 } : null,
          ]}
        >
          <Plus color={accentForeground} />
        </Pressable>
      }
    >
      <Masthead title="Chat." />
      <Typography.Paragraph color="muted">No conversations yet.</Typography.Paragraph>
    </Screen>
  );
}
