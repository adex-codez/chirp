import { Link, router } from "expo-router";
import { Button } from "heroui-native/button";
import { Card } from "heroui-native/card";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { ScrollView, TextInput, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { useSessionStore } from "@/store/use-session-store";

export default function SignUp() {
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const error = useSessionStore((state) => state.error);
  const isBusy = useSessionStore((state) => state.isBusy);
  const signUp = useSessionStore((state) => state.signUp);

  const submit = async () => {
    try {
      await signUp(username.trim(), email.trim(), password);
      router.replace("/verify");
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
            JOIN CHIRP
          </Typography>
          <Typography.Heading type="h1">Pick your Username.</Typography.Heading>
          <Typography.Paragraph color="muted">
            Your Username is unique and public. We verify your Email before you
            can use the app.
          </Typography.Paragraph>
        </View>

        <Card variant="secondary">
          <Card.Body className="gap-4">
            <View className="gap-1">
              <Typography type="body-sm" weight="bold">
                Username
              </Typography>
              <TextInput
                value={username}
                onChangeText={setUsername}
                autoCapitalize="none"
                placeholder="e.g. river_song"
                className="rounded-lg border border-separator px-3 py-2 text-foreground"
              />
            </View>
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
              {isBusy ? "Creating…" : "Sign up"}
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
