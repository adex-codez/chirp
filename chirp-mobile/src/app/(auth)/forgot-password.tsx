import { Link, router } from "expo-router";
import { Button } from "heroui-native/button";
import { Card } from "heroui-native/card";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { ScrollView, TextInput, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

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
    <SafeAreaView className="flex-1 bg-background">
      <ScrollView
        className="flex-1"
        contentContainerClassName="grow gap-6 px-6 py-8"
      >
        <View className="gap-2">
          <Typography type="body-xs" weight="bold" className="text-accent">
            RECOVER ACCESS
          </Typography>
          <Typography.Heading type="h1">Reset your password.</Typography.Heading>
          <Typography.Paragraph color="muted">
            Enter the Email of your User. If it exists and is verified, a
            code is on its way — otherwise this reveals nothing.
          </Typography.Paragraph>
        </View>

        <Card variant="secondary">
          <Card.Body className="gap-4">
            <View className="gap-1">
              <Typography type="body-sm" weight="bold">
                Email
              </Typography>
              <TextInput
                value={email}
                onChangeText={setEmail}
                autoCapitalize="none"
                keyboardType="email-address"
                editable={!isBusy}
                placeholder="you@example.com"
                className="rounded-lg border border-separator px-3 py-2 text-foreground"
              />
            </View>
            {error ? (
              <Typography.Paragraph className="text-danger">
                {error}
              </Typography.Paragraph>
            ) : null}
          </Card.Body>
          <Card.Footer className="flex-col items-stretch gap-3">
            <Button onPress={submit} isDisabled={isBusy}>
              {isBusy ? "Sending…" : "Send reset code"}
            </Button>
            <Link href="/sign-in" asChild>
              <Button variant="outline">Back to sign in</Button>
            </Link>
          </Card.Footer>
        </Card>
      </ScrollView>
    </SafeAreaView>
  );
}
