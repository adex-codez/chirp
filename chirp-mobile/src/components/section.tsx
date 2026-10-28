import type { ReactNode } from "react";
import { View } from "react-native";
import { Typography } from "heroui-native/text";

import { bandLabel } from "@/theme/typography";

type SectionProps = {
  label: string;
  children: ReactNode;
  className?: string;
};

export function Section({ label, children, className }: SectionProps) {
  return (
    <View className={`gap-3 ${className ?? ""}`}>
      <View className="flex-row items-center gap-3">
        <Typography style={bandLabel} type="body-xs" weight="bold" className="text-muted">
          {label.toUpperCase()}
        </Typography>
        <View className="h-px flex-1 bg-separator" />
      </View>
      {children}
    </View>
  );
}
