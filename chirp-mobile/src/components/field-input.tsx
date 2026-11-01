import { useState } from "react";
import {
  Pressable,
  StyleSheet,
  TextInput,
  View,
  type TextInputProps,
} from "react-native";
import { useThemeColor } from "heroui-native/hooks";
import { Typography } from "heroui-native/text";

import { dawn } from "@/theme/colors";
import { bandLabel } from "@/theme/typography";

import { Eye, EyeClosed } from "@/components/icons";

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
  secureTextEntry,
  ...rest
}: FieldInputProps) {
  const [focused, setFocused] = useState(false);
  const [revealed, setRevealed] = useState(false);
  const themePlaceholder = useThemeColor("field-placeholder");
  const muted = useThemeColor("muted");
  const withToggle = secureTextEntry === true;

  return (
    <View className="gap-2">
      <Typography style={bandLabel} type="body-xs" weight="bold" className="text-muted">
        {label.toUpperCase()}
      </Typography>
      <View style={{ position: "relative" }}>
        <TextInput
          {...rest}
          secureTextEntry={secureTextEntry && !revealed}
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
          className={`rounded-field border bg-surface py-3.5 text-base text-foreground ${
            withToggle ? "pl-4 pr-14" : "px-4"
          } ${focused ? "border-dawn" : "border-separator"} ${className ?? ""}`}
        />
        {withToggle ? (
          <Pressable
            onPress={() => setRevealed((value) => !value)}
            accessibilityRole="button"
            accessibilityLabel={revealed ? "Hide password" : "Show password"}
            style={({ pressed }) => [
              styles.toggle,
              pressed ? { opacity: 0.6 } : null,
            ]}
          >
            {revealed ? (
              <EyeClosed size={20} color={muted} />
            ) : (
              <Eye size={20} color={muted} />
            )}
          </Pressable>
        ) : null}
      </View>
      {hint ? (
        <Typography.Paragraph type="body-sm" color="muted">
          {hint}
        </Typography.Paragraph>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  toggle: {
    position: "absolute",
    top: 0,
    bottom: 0,
    right: 0,
    width: 44,
    alignItems: "center",
    justifyContent: "center",
  },
});
