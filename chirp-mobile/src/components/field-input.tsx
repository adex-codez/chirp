import { useState } from "react";
import { TextInput, View, type TextInputProps } from "react-native";
import { useThemeColor } from "heroui-native/hooks";
import { Typography } from "heroui-native/text";

import { dawn } from "@/theme/colors";
import { bandLabel } from "@/theme/typography";

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
  style,
  accessibilityLabel,
  placeholderTextColor,
  ...rest
}: FieldInputProps) {
  const [focused, setFocused] = useState(false);
  const themePlaceholder = useThemeColor("field-placeholder");

  return (
    <View className="gap-2">
      <Typography style={bandLabel} type="body-xs" weight="bold" className="text-muted">
        {label.toUpperCase()}
      </Typography>
      <TextInput
        {...rest}
        accessibilityLabel={accessibilityLabel ?? label}
        onFocus={(event) => {
          setFocused(true);
          onFocus?.(event);
        }}
        onBlur={(event) => {
          setFocused(false);
          onBlur?.(event);
        }}
        placeholderTextColor={placeholderTextColor ?? themePlaceholder}
        selectionColor={dawn.DEFAULT}
        style={[{ borderCurve: "continuous" }, style]}
        className={`rounded-field border bg-surface px-4 py-3.5 text-base text-foreground ${
          focused ? "border-dawn" : "border-separator"
        } ${className ?? ""}`}
      />
      {hint ? (
        <Typography.Paragraph type="body-sm" color="muted">
          {hint}
        </Typography.Paragraph>
      ) : null}
    </View>
  );
}
