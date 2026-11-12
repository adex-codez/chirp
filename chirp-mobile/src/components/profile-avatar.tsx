import { useState } from "react";
import { Image, StyleSheet, View, type StyleProp, type ViewStyle } from "react-native";
import { Typography } from "heroui-native/text";

import { avatarInk, spacing } from "@/theme";

export type AvatarSubject = {
  id: string;
  username: string;
  profilePictureUrl: string | null;
};

export type ProfileAvatarSize = "sm" | "md" | "lg";

export type ProfileAvatarProps = {
  subject: AvatarSubject;
  size?: ProfileAvatarSize;
  style?: StyleProp<ViewStyle>;
};

const SIZES: Record<ProfileAvatarSize, { box: number; text: "body" | "body-sm" | "body-xs" }> = {
  sm: { box: spacing.xl, text: "body-xs" },
  md: { box: spacing.xxl, text: "body-sm" },
  lg: { box: spacing.xxl + spacing.md, text: "body" },
};

function initialsFor(username: string): string {
  const segments = username
    .trim()
    .split(/[^\p{L}\p{N}]+/u)
    .filter(Boolean);
  if (segments.length === 0) return "?";

  const first = Array.from(segments[0]);
  const second = Array.from(segments[1] ?? "");
  const letters = segments.length > 1 ? [first[0], second[0]] : first.slice(0, 2);
  return letters.join("").toUpperCase() || "?";
}

export function ProfileAvatar({ subject, size = "md", style }: ProfileAvatarProps) {
  const [failedUri, setFailedUri] = useState<string | null>(null);
  const uri = subject.profilePictureUrl?.trim() ?? "";
  const showImage = uri.length > 0 && uri !== failedUri;
  const ink = avatarInk(subject.id || subject.username);
  const sizing = SIZES[size];

  return (
    <View
      accessible={false}
      accessibilityElementsHidden
      importantForAccessibility="no-hide-descendants"
      className="items-center justify-center overflow-hidden rounded-full"
      style={[{ width: sizing.box, height: sizing.box, backgroundColor: ink.background }, style]}
    >
      <Typography
        type={sizing.text}
        weight="bold"
        allowFontScaling={false}
        style={{ color: ink.foreground }}
      >
        {initialsFor(subject.username)}
      </Typography>
      {showImage ? (
        <Image
          source={{ uri }}
          style={StyleSheet.absoluteFill}
          onError={() => setFailedUri(uri)}
          accessibilityIgnoresInvertColors
        />
      ) : null}
    </View>
  );
}
