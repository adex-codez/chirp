import type { ReactNode } from "react";
import { View } from "react-native";
import { Button } from "heroui-native/button";
import { useThemeColor } from "heroui-native/hooks";
import { Typography } from "heroui-native/text";

import { ChirpLogo } from "@/components/chirp-logo";
import { ChevronLeft } from "@/components/icons";
import { bandLabel, displayTitle } from "@/theme/typography";

type MastheadProps = {
  title: string;
  /** Printed verbatim: pass the casing the value actually has, since
   * Usernames are case-sensitive. */
  meta?: string;
  onBack?: () => void;
  children?: ReactNode;
};

export function Masthead({ title, meta, onBack, children }: MastheadProps) {
  const muted = useThemeColor("muted");

  return (
    <View className="gap-4">
      <View className="flex-row items-center justify-between">
        {onBack ? (
          <Button
            variant="ghost"
            size="sm"
            isIconOnly
            onPress={onBack}
            accessibilityLabel="Go back"
            hitSlop={8}
            className="-ml-3"
          >
            <ChevronLeft color={muted} />
          </Button>
        ) : (
          <View className="flex-row items-center gap-2.5">
            <ChirpLogo size={28} />
            <Typography style={bandLabel} type="body-xs" weight="bold">
              CHIRP
            </Typography>
          </View>
        )}
        {meta ? (
          <Typography style={bandLabel} type="body-xs" className="text-muted">
            {meta}
          </Typography>
        ) : null}
      </View>
      <View className="h-px bg-separator" />
      <Typography.Heading style={displayTitle} type="h2">
        {title}
      </Typography.Heading>
      {children ? <View className="gap-3">{children}</View> : null}
    </View>
  );
}
