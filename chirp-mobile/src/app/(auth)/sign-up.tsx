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
    } catch {
      // Error text already lives in the store.
    }
  };

  return (
    <AuthShell
      kicker="BANDING STATION"
      title="Claim your call sign."
      intro="Your Username is unique and public. We check your Email before anything else unlocks."
      step={1}
    >
      <ScrollView className="flex-1" contentContainerClassName="grow gap-5 pb-8">
        <View className="gap-4">
          <FieldInput
            label="Username"
            value={username}
            onChangeText={setUsername}
            autoCapitalize="none"
            placeholder="e.g. river_song"
            hint="3–20 characters: letters, numbers, underscore, dot."
          />
          <FieldInput
            label="Email"
            value={email}
            onChangeText={setEmail}
            autoCapitalize="none"
            keyboardType="email-address"
            placeholder="you@example.com"
            hint="We send one code here. Nothing else happens until you enter it."
          />
          <FieldInput
            label="Password"
            value={password}
            onChangeText={setPassword}
            secureTextEntry
            placeholder="8-15 chars, upper, lower, number & symbol"
          />
          {error ? (
            <Typography.Paragraph className="text-danger">{error}</Typography.Paragraph>
          ) : null}
        </View>

        <Button onPress={submit} isDisabled={isBusy}>
          {isBusy ? "Creating…" : "Claim and continue"}
        </Button>

        <View className="flex-row items-center gap-3">
          <View className="h-px flex-1 bg-separator" />
          <Songline height={20} levels={[10, 22, 14, 30, 18, 26, 12]} />
          <View className="h-px flex-1 bg-separator" />
        </View>

        <SocialButtons />

        <View className="flex-row items-center justify-center gap-1.5">
          <Typography color="muted">Have an account?</Typography>
          <Link href="/sign-in" asChild>
            <Typography weight="bold" className="text-accent">
              Sign in
            </Typography>
          </Link>
        </View>
      </ScrollView>
    </AuthShell>
  );
}
