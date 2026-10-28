import { Link, router } from "expo-router";
import { Button } from "heroui-native/button";
import { InputOTP, REGEXP_ONLY_DIGITS } from "heroui-native/input-otp";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { View } from "react-native";

import { AuthShell } from "@/components/auth-shell";
import { FieldInput } from "@/components/field-input";
import { Notice } from "@/components/notice";
import { useSessionStore } from "@/store/use-session-store";
import { bandLabel } from "@/theme/typography";

export default function ResetPassword() {
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const error = useSessionStore((state) => state.error);
  const isBusy = useSessionStore((state) => state.isBusy);
  const resetPassword = useSessionStore((state) => state.resetPassword);

  const submit = async () => {
    try {
      await resetPassword(email.trim(), code.trim(), newPassword);
      router.replace("/sign-in");
    } catch {
      // Error text already lives in the store.
    }
  };

  return (
    <AuthShell
      kicker="NEW SECRET"
      title="Choose a password."
      intro="8–15 characters with an uppercase letter, a lowercase letter, a number, and a special character. Resetting signs every device out."
      footer={
        <Link href="/sign-in" asChild>
          <Button variant="ghost">Back to sign in</Button>
        </Link>
      }
    >
      <View className="gap-4">
        <FieldInput
          label="Email"
          value={email}
          onChangeText={setEmail}
          autoCapitalize="none"
          keyboardType="email-address"
          editable={!isBusy}
          placeholder="you@example.com"
        />
        <View className="gap-2">
          <Typography style={bandLabel} type="body-xs" weight="bold" className="text-muted">
            CODE
          </Typography>
          <InputOTP
            value={code}
            onChange={setCode}
            maxLength={6}
            isInvalid={Boolean(error)}
            isDisabled={isBusy}
            pattern={REGEXP_ONLY_DIGITS}
            textInputProps={{ accessibilityLabel: "Six digit reset code" }}
          >
            <InputOTP.Group>
              <InputOTP.Slot index={0} className="w-10" />
              <InputOTP.Slot index={1} className="w-10" />
              <InputOTP.Slot index={2} className="w-10" />
            </InputOTP.Group>
            <InputOTP.Separator />
            <InputOTP.Group>
              <InputOTP.Slot index={3} className="w-10" />
              <InputOTP.Slot index={4} className="w-10" />
              <InputOTP.Slot index={5} className="w-10" />
            </InputOTP.Group>
          </InputOTP>
        </View>
        <FieldInput
          label="New password"
          value={newPassword}
          onChangeText={setNewPassword}
          secureTextEntry
          editable={!isBusy}
          placeholder="8-15 chars, upper, lower, number & symbol"
        />
        {error ? <Notice tone="danger">{error}</Notice> : null}
      </View>

      <Button onPress={submit} isDisabled={isBusy}>
        {isBusy ? "Resetting…" : "Reset password"}
      </Button>
    </AuthShell>
  );
}
