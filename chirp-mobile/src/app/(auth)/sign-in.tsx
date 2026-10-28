import { Link } from "expo-router";
import { Button } from "heroui-native/button";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { View } from "react-native";

import { AuthShell } from "@/components/auth-shell";
import { FieldInput } from "@/components/field-input";
import { Notice } from "@/components/notice";
import { SocialButtons } from "@/components/social-buttons";
import { SonglineRule } from "@/components/songline";
import { useSessionStore } from "@/store/use-session-store";

export default function SignIn() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const error = useSessionStore((state) => state.error);
  const isBusy = useSessionStore((state) => state.isBusy);
  const signIn = useSessionStore((state) => state.signIn);
  const status = useSessionStore((state) => state.status);

  const submit = async () => {
    try {
      await signIn(email.trim(), password);
    } catch {
      // Error text already lives in the store.
    }
  };

  return (
    <AuthShell
      title="Welcome back."
      intro="Sign in to pick up your conversations."
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
        <FieldInput
          label="Password"
          value={password}
          onChangeText={setPassword}
          secureTextEntry
          editable={!isBusy}
          placeholder="Your password"
        />
        {error ? <Notice tone="danger">{error}</Notice> : null}
        {status === "pending-verification" ? (
          <Typography.Paragraph color="muted">
            That Email is not verified yet — check your inbox for the code.
          </Typography.Paragraph>
        ) : null}
      </View>

      <Button onPress={submit} isDisabled={isBusy}>
        {isBusy ? "Signing in…" : "Sign in"}
      </Button>

      <SonglineRule />

      <SocialButtons />

      <View className="items-center gap-1">
        <Link href="/forgot-password" asChild>
          <Button variant="ghost">Forgot password?</Button>
        </Link>
        <View className="flex-row items-center justify-center gap-1.5">
          <Typography color="muted">Don&apos;t have an account?</Typography>
          <Link href="/sign-up" asChild>
            <Typography accessibilityRole="link" weight="bold" className="py-3 text-accent">
              Sign up
            </Typography>
          </Link>
        </View>
      </View>
    </AuthShell>
  );
}
