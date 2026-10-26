import { useState } from "react";
import { TextInput, View, type TextInputProps } from "react-native";
import { useThemeColor } from "heroui-native/hooks";
import { Typography } from "heroui-native/text";
import { fontStack } from "@/theme/typography";

/*
 * FieldInput — mono band label over a large field.
 * Labels name what the User controls (Username, Email), never
 * system terms. Gold focus ring ties every field to the songline.
 */

type FieldInputProps = TextInputProps & {
  label: string;
  hint?: string;
};

export function FieldInput({
  label,
  hint,
  onFocus,
  onBlur,
  className,
  accessibilityLabel,
  ...rest
}: FieldInputProps) {
  const [focused, setFocused] = useState(false);
  const placeholderColor = useThemeColor("field-placeholder");
  return (
    <View className="gap-1.5">
      <Typography
        style={{ fontFamily: fontStack.band, letterSpacing: 1.5 }}
        type="body-xs"
        weight="bold"
        className="text-foreground"
      >
        {label.toUpperCase()}
      </Typography>
      <TextInput
        {...rest}
        accessibilityLabel={accessibilityLabel ?? label}
        onFocus={(e) => {
          setFocused(true);
          onFocus?.(e);
        }}
        onBlur={(e) => {
          setFocused(false);
          onBlur?.(e);
        }}
        placeholderTextColor={placeholderColor}
        className={`rounded-field border-2 bg-surface px-4 py-3 text-base text-foreground ${
          focused ? "border-dawn" : "border-separator"
        } ${className ?? ""}`}
      />
      {hint ? <Typography.Paragraph color="muted">{hint}</Typography.Paragraph> : null}
    </View>
  );
}
