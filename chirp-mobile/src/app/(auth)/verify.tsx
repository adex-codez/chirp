import { Button } from "heroui-native/button";
import { InputOTP, REGEXP_ONLY_DIGITS } from "heroui-native/input-otp";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { View } from "react-native";

import { AuthShell } from "@/components/auth-shell";
import { Notice } from "@/components/notice";
import { useSessionStore } from "@/store/use-session-store";
import { bandLabel } from "@/theme/typography";

export default function Verify() {
  const [code, setCode] = useState("");
  const pendingEmail = useSessionStore((state) => state.pendingEmail);
  const error = useSessionStore((state) => state.error);
  const isBusy = useSessionStore((state) => state.isBusy);
  const verify = useSessionStore((state) => state.verify);
  const resendCode = useSessionStore((state) => state.resendCode);

  const submit = async (value: string) => {
    if (value.length !== 6) return;
    try {
      await verify(value);
    } catch {
      // Error text already lives in the store.
    }
  };

  return (
    <AuthShell
      title="Check your email."
      intro={
        pendingEmail
          ? `We sent a 6-digit code to ${pendingEmail}. It expires in 20 minutes.`
          : "Choose a Username first, then enter the code we send you."
      }
      step={2}
      footer={
        <Button variant="ghost" onPress={() => void resendCode()} isDisabled={isBusy}>
          Resend code
        </Button>
      }
    >
      <View className="gap-4">
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
            onComplete={(value) => void submit(value)}
            textInputProps={{ accessibilityLabel: "Six digit verification code" }}
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
        <Typography.Paragraph type="body-sm" color="muted">
          Enter the six-digit code from the email.
        </Typography.Paragraph>
        {error ? <Notice tone="danger">{error}</Notice> : null}
      </View>

      <Button onPress={() => void submit(code)} isDisabled={isBusy}>
        {isBusy ? "Verifying…" : "Verify and continue"}
      </Button>
    </AuthShell>
  );
}
