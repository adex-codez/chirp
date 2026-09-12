import { Link, router } from "expo-router";
import { Button } from "heroui-native/button";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { ScrollView, View } from "react-native";
import { AuthShell } from "@/components/auth-shell";
import { FieldInput } from "@/components/field-input";

import { useSessionStore } from "@/store/use-session-store";

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
    >
      <ScrollView className="flex-1" contentContainerClassName="grow gap-5 pb-8">
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
          <FieldInput
            label="Code"
            value={code}
            onChangeText={setCode}
            keyboardType="number-pad"
            maxLength={6}
            editable={!isBusy}
            placeholder="123456"
          />
          <FieldInput
            label="New password"
            value={newPassword}
            onChangeText={setNewPassword}
            secureTextEntry
            editable={!isBusy}
            placeholder="8-15 chars, upper, lower, number & symbol"
          />
          {error ? (
            <Typography.Paragraph className="text-danger">{error}</Typography.Paragraph>
          ) : null}
        </View>
        <View className="gap-3">
          <Button onPress={submit} isDisabled={isBusy}>
            {isBusy ? "Resetting…" : "Reset password"}
          </Button>
          <Link href="/sign-in" asChild>
            <Button variant="ghost">Back to sign in</Button>
          </Link>
        </View>
      </ScrollView>
    </AuthShell>
  );
}
