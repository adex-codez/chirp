import { useRouter } from "expo-router";
import { Button } from "heroui-native/button";
import { Typography } from "heroui-native/text";
import { useState } from "react";

import { BandChip } from "@/components/band-chip";
import { FieldInput } from "@/components/field-input";
import { Masthead } from "@/components/masthead";
import { Notice } from "@/components/notice";
import { Screen } from "@/components/screen";
import { Section } from "@/components/section";
import { useSessionStore } from "@/store/use-session-store";

export default function Settings() {
  const router = useRouter();
  const user = useSessionStore((state) => state.user);
  const error = useSessionStore((state) => state.error);
  const isBusy = useSessionStore((state) => state.isBusy);
  const changeUsername = useSessionStore((state) => state.changeUsername);
  const addPassword = useSessionStore((state) => state.addPassword);
  const signOutEverywhere = useSessionStore((state) => state.signOutEverywhere);

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
    <Screen>
      <Masthead
        title="Settings."
        onBack={() => (router.canGoBack() ? router.back() : router.replace("/"))}
      >
        {user ? <BandChip username={user.username} verified={user.verified} /> : null}
      </Masthead>

      <Section label="Username">
        <FieldInput
          label="Username"
          value={username}
          onChangeText={setUsernameDraft}
          autoCapitalize="none"
          editable={!isBusy}
          placeholder="e.g. river_song"
          hint="3–20 characters: letters, numbers, underscore, dot."
        />
        <Button onPress={submitUsername} isDisabled={isBusy}>
          {isBusy ? "Saving…" : "Save Username"}
        </Button>
      </Section>

      {user && !user.hasPassword ? (
        <Section label="Password">
          <Typography.Paragraph color="muted">
            This User joined with Social sign-in and has no password yet. Adding one unlocks
            Password sign-in too.
          </Typography.Paragraph>
          <FieldInput
            label="New password"
            value={newPassword}
            onChangeText={setNewPassword}
            secureTextEntry
            editable={!isBusy}
            placeholder="8-15 chars, upper, lower, number & symbol"
          />
          <Button variant="secondary" onPress={submitPassword} isDisabled={isBusy}>
            {isBusy ? "Saving…" : "Add password"}
          </Button>
        </Section>
      ) : null}

      <Section label="Devices">
        <Typography.Paragraph color="muted">
          {user
            ? `Signed in as ${user.email} on this device. Signing out everywhere ends every other session.`
            : "Signing out everywhere ends every other session."}
        </Typography.Paragraph>
        <Button variant="outline" onPress={() => void signOutEverywhere()} isDisabled={isBusy}>
          Sign out everywhere
        </Button>
      </Section>

      {saved ? <Notice tone="success">{saved}</Notice> : null}
      {error ? <Notice tone="danger">{error}</Notice> : null}
    </Screen>
  );
}
