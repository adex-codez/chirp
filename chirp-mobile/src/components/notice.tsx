import type { ReactNode } from "react";
import { View } from "react-native";
import { Typography } from "heroui-native/text";

type NoticeTone = "danger" | "success" | "muted";

type NoticeProps = {
  tone: NoticeTone;
  children: ReactNode;
};

const TONES: Record<NoticeTone, { box: string; text: string }> = {
  danger: { box: "border-danger bg-danger-soft", text: "text-danger-soft-foreground" },
  success: { box: "border-success bg-success-soft", text: "text-success-soft-foreground" },
  muted: { box: "border-separator bg-surface", text: "text-muted" },
};

export function Notice({ tone, children }: NoticeProps) {
  const { box, text } = TONES[tone];

  return (
    <View
      className={`rounded-xl border-l-2 px-3 py-2.5 ${box}`}
      style={{ borderCurve: "continuous" }}
    >
      <Typography type="body-sm" className={text}>
        {children}
      </Typography>
    </View>
  );
}
