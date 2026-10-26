import { Link, router } from "expo-router";
import { Button } from "heroui-native/button";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { ScrollView, View } from "react-native";
import { AuthShell } from "@/components/auth-shell";
import { FieldInput } from "@/components/field-input";

import { useSessionStore } from "@/store/use-session-store";

export default function ForgotPassword() {
  const [email, setEmail] = useState("");
  const error = useSessionStore((state) => state.error);
  const isBusy = useSessionStore((state) => state.isBusy);
  const forgotPassword = useSessionStore((state) => state.forgotPassword);

  const submit = async () => {
    try {
      await forgotPassword(email.trim());
      router.replace("/reset-password");
    } catch {
      // Error text already lives in the store.
    }
  };

  return (
    <AuthShell
      kicker="LOST YOUR SONG?"
      title="Find your way back."
      intro="Enter the Email of your User. If it exists and is verified, a code is on its way — otherwise this reveals nothing."
    >
      <ScrollView className="flex-1" contentContainerClassName="grow gap-5 pb-8">
        <FieldInput
          label="Email"
          value={email}
          onChangeText={setEmail}
          autoCapitalize="none"
          keyboardType="email-address"
          editable={!isBusy}
          placeholder="you@example.com"
        />
        {error ? (
          <Typography.Paragraph className="text-danger">{error}</Typography.Paragraph>
        ) : null}
        <View className="gap-3">
          <Button onPress={submit} isDisabled={isBusy}>
            {isBusy ? "Sending…" : "Send reset code"}
          </Button>
          <Link href="/sign-in" asChild>
            <Button variant="ghost">Back to sign in</Button>
          </Link>
        </View>
      </ScrollView>
    </AuthShell>
  );
}
