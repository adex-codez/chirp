import { router } from "expo-router";
import { Button } from "heroui-native/button";
import { Card } from "heroui-native/card";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { ScrollView, TextInput, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { useSessionStore } from "@/store/use-session-store";

export default function Verify() {
  const [code, setCode] = useState("");
  const pendingEmail = useSessionStore((state) => state.pendingEmail);
  const error = useSessionStore((state) => state.error);
  const isBusy = useSessionStore((state) => state.isBusy);
  const verify = useSessionStore((state) => state.verify);
  const resendCode = useSessionStore((state) => state.resendCode);

  const submit = async () => {
    try {
      await verify(code.trim());
      router.replace("/");
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
            CHECK YOUR INBOX
          </Typography>
          <Typography.Heading type="h1">Enter the code.</Typography.Heading>
          <Typography.Paragraph color="muted">
            {pendingEmail
              ? `We sent a 6-digit code to ${pendingEmail}. It expires in about 20 minutes.`
              : "Create an account first, then enter the code we send you."}
          </Typography.Paragraph>
        </View>

        <Card variant="secondary">
          <Card.Body className="gap-4">
            <View className="gap-1">
              <Typography type="body-sm" weight="bold">
                Code
              </Typography>
              <TextInput
                value={code}
                onChangeText={setCode}
                keyboardType="number-pad"
                maxLength={6}
                placeholder="123456"
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
              {isBusy ? "Verifying…" : "Verify and continue"}
            </Button>
            <Button
              variant="outline"
              onPress={() => resendCode()}
              isDisabled={isBusy}
            >
              Resend code
            </Button>
          </Card.Footer>
        </Card>
      </ScrollView>
    </SafeAreaView>
  );
}
