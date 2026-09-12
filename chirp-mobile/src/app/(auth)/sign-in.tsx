import { Link } from "expo-router";
import { Button } from "heroui-native/button";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { ScrollView, View } from "react-native";
import { AuthShell } from "@/components/auth-shell";
import { FieldInput } from "@/components/field-input";
import { Songline } from "@/components/songline";
import { SocialButtons } from "@/components/social-buttons";

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
      kicker="RETURNING CHORUS"
      title="Welcome back."
      intro="Sign in with your Email and password, or use Social sign-in."
    >
      <ScrollView className="flex-1" contentContainerClassName="grow gap-4 pb-6">
        <View className="gap-4">
          <FieldInput
            label="Email"
            value={email}
            onChangeText={setEmail}
            autoCapitalize="none"
            keyboardType="email-address"
            placeholder="you@example.com"
          />
          <FieldInput
            label="Password"
            value={password}
            onChangeText={setPassword}
            secureTextEntry
            placeholder="Your password"
          />
          {error ? (
            <Typography.Paragraph className="text-danger">{error}</Typography.Paragraph>
          ) : null}
          {status === "pending-verification" ? (
            <Typography.Paragraph color="muted">
              That Email is not verified yet — check your inbox for the code.
            </Typography.Paragraph>
          ) : null}
        </View>

        <Button onPress={submit} isDisabled={isBusy}>
          {isBusy ? "Signing in…" : "Sign in"}
        </Button>

        <View className="flex-row items-center gap-3">
          <View className="h-px flex-1 bg-separator" />
          <Songline height={20} levels={[10, 22, 14, 30, 18, 26, 12]} />
          <View className="h-px flex-1 bg-separator" />
        </View>

        <SocialButtons />

        <View className="items-center gap-0.5">
          <Link href="/forgot-password" asChild>
            <Button variant="ghost" size="sm">
              Forgot password?
            </Button>
          </Link>
          <View className="flex-row items-center justify-center gap-1.5">
            <Typography color="muted">Don&apos;t have an account?</Typography>
            <Link href="/sign-up" asChild>
              <Typography weight="bold" className="text-accent">
                Sign up
              </Typography>
            </Link>
          </View>
        </View>
      </ScrollView>
    </AuthShell>
  );
}
