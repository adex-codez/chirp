import { Link, router } from "expo-router";
import { Button } from "heroui-native/button";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { View } from "react-native";

import { AuthShell } from "@/components/auth-shell";
import { CodeField } from "@/components/code-field";
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
  const clearError = useSessionStore((state) => state.clearError);

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
      title="Set a new password."
      intro="8–15 characters, with an uppercase letter, a lowercase letter, a number and a symbol. Resetting signs you out on every device."
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
          <CodeField
            value={code}
            onChange={setCode}
            onEdit={error ? clearError : undefined}
            isInvalid={Boolean(error)}
            isDisabled={isBusy}
            label="Six digit reset code"
          />
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
