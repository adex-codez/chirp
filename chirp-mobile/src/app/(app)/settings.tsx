import { Button } from "heroui-native/button";
import { Card } from "heroui-native/card";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { ScrollView, TextInput, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { useSessionStore } from "@/store/use-session-store";

export default function Settings() {
  const user = useSessionStore((state) => state.user);
  const error = useSessionStore((state) => state.error);
  const isBusy = useSessionStore((state) => state.isBusy);
  const changeUsername = useSessionStore((state) => state.changeUsername);
  const addPassword = useSessionStore((state) => state.addPassword);
  const signOutEverywhere = useSessionStore(
    (state) => state.signOutEverywhere,
  );

  const [username, setUsernameDraft] = useState(user?.username ?? "");
  const [newPassword, setNewPassword] = useState("");
  const [saved, setSaved] = useState<string | null>(null);

  const submitUsername = async () => {
    setSaved(null);
    try {
      await changeUsername(username.trim());
      setSaved("Username updated.");
    } catch {
      // Error text already lives in the store.
    }
  };

  const submitPassword = async () => {
    setSaved(null);
    try {
      await addPassword(newPassword);
      setNewPassword("");
      setSaved("Password added. Both sign-in methods reach this User now.");
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
            YOUR USER
          </Typography>
          <Typography.Heading type="h1">Settings.</Typography.Heading>
          <Typography.Paragraph color="muted">
            {user
              ? `Signed in as ${user.username} (${user.email}).`
              : "Manage this User."}
          </Typography.Paragraph>
        </View>

        <Card variant="secondary">
          <Card.Body className="gap-4">
            <Typography.Heading type="h3">Username</Typography.Heading>
            <View className="gap-1">
              <TextInput
                value={username}
                onChangeText={setUsernameDraft}
                autoCapitalize="none"
                editable={!isBusy}
                placeholder="e.g. river_song"
                className="rounded-lg border border-separator px-3 py-2 text-foreground"
              />
              <Typography.Paragraph color="muted">
                3–20 characters: letters, numbers, underscore, dot.
              </Typography.Paragraph>
            </View>
            <Button onPress={submitUsername} isDisabled={isBusy}>
              {isBusy ? "Saving…" : "Save Username"}
            </Button>
          </Card.Body>
        </Card>

        {user && !user.hasPassword ? (
          <Card variant="secondary">
            <Card.Body className="gap-4">
              <Typography.Heading type="h3">Add password</Typography.Heading>
              <Typography.Paragraph color="muted">
                This User joined with Social sign-in and has no password
                yet. Adding one unlocks Password sign-in too.
              </Typography.Paragraph>
              <View className="gap-1">
                <TextInput
                  value={newPassword}
                  onChangeText={setNewPassword}
                  secureTextEntry
                  editable={!isBusy}
                  placeholder="8-15 chars, upper, lower, number & symbol"
                  className="rounded-lg border border-separator px-3 py-2 text-foreground"
                />
              </View>
              <Button onPress={submitPassword} isDisabled={isBusy}>
                {isBusy ? "Saving…" : "Add password"}
              </Button>
            </Card.Body>
          </Card>
        ) : null}

        <Card variant="secondary">
          <Card.Body className="gap-4">
            <Typography.Heading type="h3">Devices</Typography.Heading>
            <Button
              variant="outline"
              onPress={() => void signOutEverywhere()}
              isDisabled={isBusy}
            >
              Sign out everywhere
            </Button>
          </Card.Body>
        </Card>

        {saved ? (
          <Typography.Paragraph color="muted">{saved}</Typography.Paragraph>
        ) : null}
        {error ? (
          <Typography.Paragraph className="text-danger">
            {error}
          </Typography.Paragraph>
        ) : null}
      </ScrollView>
    </SafeAreaView>
  );
}
