import { Link, router } from "expo-router";
import { Button } from "heroui-native/button";
import { Card } from "heroui-native/card";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { ScrollView, TextInput, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

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
    <SafeAreaView className="flex-1 bg-background">
      <ScrollView
        className="flex-1"
        contentContainerClassName="grow gap-6 px-6 py-8"
      >
        <View className="gap-2">
          <Typography type="body-xs" weight="bold" className="text-accent">
            NEW SECRET
          </Typography>
          <Typography.Heading type="h1">Choose a password.</Typography.Heading>
          <Typography.Paragraph color="muted">
            8–15 characters with an uppercase letter, a lowercase letter, a
            number, and a special character. Resetting signs every device
            out.
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
            <View className="gap-1">
              <Typography type="body-sm" weight="bold">
                Code
              </Typography>
              <TextInput
                value={code}
                onChangeText={setCode}
                keyboardType="number-pad"
                maxLength={6}
                editable={!isBusy}
                placeholder="123456"
                className="rounded-lg border border-separator px-3 py-2 text-foreground"
              />
            </View>
            <View className="gap-1">
              <Typography type="body-sm" weight="bold">
                New password
              </Typography>
              <TextInput
                value={newPassword}
                onChangeText={setNewPassword}
                secureTextEntry
                editable={!isBusy}
                placeholder="8-15 chars, upper, lower, number & symbol"
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
              {isBusy ? "Resetting…" : "Reset password"}
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
