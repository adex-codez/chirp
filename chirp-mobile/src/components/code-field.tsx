import { useEffect, useRef, useState } from "react";
import { Animated, Platform, TextInput, View } from "react-native";

import { dawn } from "@/theme/colors";
import { motion } from "@/theme/motion";

type CodeFieldProps = {
  value: string;
  onChange: (value: string) => void;
  onComplete?: (value: string) => void;
  onEdit?: () => void;
  length?: number;
  isInvalid?: boolean;
  isDisabled?: boolean;
  label: string;
};

const SHAKE_OFFSET = 7;
const SHAKE_STEP = motion.fast / 3;

export function CodeField({
  value,
  onChange,
  onComplete,
  onEdit,
  length = 6,
  isInvalid = false,
  isDisabled = false,
  label,
}: CodeFieldProps) {
  const inputs = useRef<Array<TextInput | null>>([]);
  const shift = useRef(new Animated.Value(0)).current;
  const wasInvalid = useRef(isInvalid);
  const [focusedIndex, setFocusedIndex] = useState<number | null>(null);

  const chars = Array.from({ length }, (_, index) => value[index] ?? "");
  const half = Math.ceil(length / 2);

  useEffect(() => {
    if (isInvalid && !wasInvalid.current) {
      Animated.sequence(
        [SHAKE_OFFSET, -SHAKE_OFFSET, SHAKE_OFFSET / 2, 0].map((toValue) =>
          Animated.timing(shift, {
            toValue,
            duration: SHAKE_STEP,
            useNativeDriver: Platform.OS !== "web",
          }),
        ),
      ).start();
    }
    wasInvalid.current = isInvalid;
  }, [isInvalid, shift]);

  const focusSlot = (index: number) => {
    inputs.current[Math.max(0, Math.min(index, length - 1))]?.focus();
  };

  const commit = (next: string) => {
    onChange(next);
    if (next.length === length) onComplete?.(next);
  };

  const handleChange = (index: number, text: string) => {
    onEdit?.();
    const digits = text.replace(/\D/g, "");
    if (!digits) {
      commit(value.slice(0, index) + value.slice(index + 1));
      return;
    }
    const inserted = digits.slice(0, length - index);
    commit((value.slice(0, index) + inserted + value.slice(index + inserted.length)).slice(0, length));
    focusSlot(index + inserted.length);
  };

  const handleBackspace = (index: number, key: string) => {
    if (key !== "Backspace" || chars[index] || index === 0) return;
    onEdit?.();
    focusSlot(index - 1);
    commit(value.slice(0, index - 1) + value.slice(index));
  };

  const slot = (index: number) => (
    <TextInput
      key={index}
      ref={(node) => {
        inputs.current[index] = node;
      }}
      value={chars[index]}
      onChangeText={(text) => handleChange(index, text)}
      onKeyPress={({ nativeEvent }) => handleBackspace(index, nativeEvent.key)}
      onFocus={() => setFocusedIndex(index)}
      onBlur={() => setFocusedIndex((current) => (current === index ? null : current))}
      selection={focusedIndex === index && chars[index] ? { start: 0, end: 1 } : undefined}
      maxLength={length}
      editable={!isDisabled}
      selectionColor={dawn.DEFAULT}
      keyboardType="number-pad"
      autoComplete={index === 0 ? "one-time-code" : "off"}
      textContentType={index === 0 ? "oneTimeCode" : "none"}
      accessibilityLabel={`${label}, digit ${index + 1}`}
      accessibilityValue={{ text: chars[index] || "empty" }}
      style={{ borderCurve: "continuous" }}
      className={`h-12 w-10 rounded-field border bg-surface text-center text-lg font-semibold text-foreground ${
        isInvalid ? "border-danger" : focusedIndex === index ? "border-dawn" : "border-separator"
      } ${isDisabled ? "opacity-50" : ""}`}
    />
  );

  return (
    <Animated.View style={{ transform: [{ translateX: shift }] }}>
      <View className="flex-row items-center gap-2">
        {Array.from({ length: half }, (_, index) => slot(index))}
        <View className="h-0.5 w-2 rounded-full bg-separator" />
        {Array.from({ length: length - half }, (_, index) => slot(half + index))}
      </View>
    </Animated.View>
  );
}
