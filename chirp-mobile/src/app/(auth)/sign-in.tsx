import { Link } from "expo-router";
import { Button } from "heroui-native/button";
import { Card } from "heroui-native/card";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { ScrollView, TextInput, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { useSessionStore } from "@/store/use-session-store";
import { SocialButtons } from "@/components/social-buttons";

export default function SignIn() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const error = useSessionStore((state) => state.error);
  const isBusy = useSessionStore((state) => state.isBusy);
  const signIn = useSessionStore((state) => state.signIn);
  const status = useSessionStore((state) => state.status);

  // Transitions are owned by the route guards: success flips the session
  // to authenticated and the protected group takes over; unverified
  // sign-ins park at pending-verification and the root layout sends them
  // to verify. Errors surface from the store.
  const submit = async () => {
    try {
      await signIn(email.trim(), password);
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
            WELCOME BACK
          </Typography>
          <Typography.Heading type="h1">Sign in to Chirp.</Typography.Heading>
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
                placeholder="you@example.com"
                className="rounded-lg border border-separator px-3 py-2 text-foreground"
              />
            </View>
            <View className="gap-1">
              <Typography type="body-sm" weight="bold">
                Password
              </Typography>
              <TextInput
                value={password}
                onChangeText={setPassword}
                secureTextEntry
                placeholder="Your password"
                className="rounded-lg border border-separator px-3 py-2 text-foreground"
              />
            </View>
            {error ? (
              <Typography.Paragraph className="text-danger">
                {error}
              </Typography.Paragraph>
            ) : null}
            {status === "pending-verification" ? (
              <Typography.Paragraph color="muted">
                That Email is not verified yet — check your inbox for the code.
              </Typography.Paragraph>
            ) : null}
          </Card.Body>
          <Card.Footer className="flex-col items-stretch gap-3">
            <Button onPress={submit} isDisabled={isBusy}>
              {isBusy ? "Signing in…" : "Sign in"}
            </Button>
            <SocialButtons />
            <Link href="/sign-up" asChild>
              <Button variant="outline">Sign up instead</Button>
            </Link>
          </Card.Footer>
        </Card>
      </ScrollView>
    </SafeAreaView>
  );
}
