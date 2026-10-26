import { Button } from "heroui-native/button";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { ScrollView, View } from "react-native";
import { AuthShell } from "@/components/auth-shell";
import { FieldInput } from "@/components/field-input";

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
    } catch {
      // Error text already lives in the store.
    }
  };

  return (
    <AuthShell
      kicker="CHECK YOUR INBOX"
      title="Listen for the code."
      intro={
        pendingEmail
          ? `We sent a 6-digit code to ${pendingEmail}. It fades in about 20 minutes.`
          : "Claim a Username first, then enter the code we send you."
      }
      step={2}
    >
      <ScrollView className="flex-1" contentContainerClassName="grow gap-5 pb-8">
        <FieldInput
          label="Code"
          value={code}
          onChangeText={setCode}
          keyboardType="number-pad"
          maxLength={6}
          placeholder="123456"
          hint="Six digits. Keep this screen open while you check your inbox."
        />
        {error ? (
          <Typography.Paragraph className="text-danger">{error}</Typography.Paragraph>
        ) : null}
        <View className="gap-3">
          <Button onPress={submit} isDisabled={isBusy}>
            {isBusy ? "Verifying…" : "Verify and continue"}
          </Button>
          <Button variant="outline" onPress={() => resendCode()} isDisabled={isBusy}>
            Resend code
          </Button>
        </View>
      </ScrollView>
    </AuthShell>
  );
}
