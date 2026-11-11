import { Button } from "@/components/button";
import { Avatar } from "heroui-native/avatar";
import { ControlField } from "heroui-native/control-field";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { View } from "react-native";

import { BandChip } from "@/components/band-chip";
import { FieldInput } from "@/components/field-input";
import { Masthead } from "@/components/masthead";
import { Notice } from "@/components/notice";
import { Screen } from "@/components/screen";
import { Section } from "@/components/section";
import { useTabBarSpace } from "@/components/tab-bar";
import { useAppStore } from "@/store/use-app-store";
import { useSessionStore } from "@/store/use-session-store";

export default function Settings() {
  const barSpace = useTabBarSpace();
  const user = useSessionStore((state) => state.user);
  const error = useSessionStore((state) => state.error);
  const isBusy = useSessionStore((state) => state.isBusy);
  const pictureState = useSessionStore((state) => state.pictureState);
  const changeProfilePicture = useSessionStore((state) => state.changeProfilePicture);
  const changeUsername = useSessionStore((state) => state.changeUsername);
  const addPassword = useSessionStore((state) => state.addPassword);
  const signOutEverywhere = useSessionStore((state) => state.signOutEverywhere);
  const signOut = useSessionStore((state) => state.signOut);
  const isNotificationsEnabled = useAppStore((state) => state.isNotificationsEnabled);
  const toggleNotifications = useAppStore((state) => state.toggleNotifications);

  const [username, setUsernameDraft] = useState(user?.username ?? "");
  const [newPassword, setNewPassword] = useState("");
  const [saved, setSaved] = useState<string | null>(null);

  const canSubmitUsername = username.trim().length > 0;
  const canSubmitPassword = newPassword.trim().length > 0;
  const accountBusy = isBusy || pictureState.kind === "changing";

  const submitPicture = async () => {
    if (accountBusy) return;
    setSaved(null);
    try {
      if (await changeProfilePicture() === "saved") {
        setSaved("Profile picture updated.");
      }
    } catch {}
  };

  const submitUsername = async () => {
    if (accountBusy || !canSubmitUsername) return;
    setSaved(null);
    try {
      await changeUsername(username.trim());
      setSaved("Username updated.");
    } catch {
      // Error text already lives in the store.
    }
  };

  const submitPassword = async () => {
    if (accountBusy || !canSubmitPassword) return;
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
    <Screen bottomSpace={barSpace}>
      <Masthead title="Settings.">
        {user ? <BandChip username={user.username} verified={user.verified} /> : null}
      </Masthead>

      <Section label="Account">
        <View className="flex-row items-center gap-4">
          <Avatar size="lg" color="default" alt="Your profile picture" animation="disable-all">
            {user?.profilePictureUrl ? (
              <Avatar.Image key={user.profilePictureUrl} source={{ uri: user.profilePictureUrl }} />
            ) : null}
            <Avatar.Fallback>{user?.username.slice(0, 1).toUpperCase() || "C"}</Avatar.Fallback>
          </Avatar>
          <View className="flex-1">
            <Typography>Profile picture</Typography>
            <Typography.Paragraph type="body-sm" color="muted">
              Choose a photo from your library.
            </Typography.Paragraph>
          </View>
        </View>
        <Button variant="secondary" onPress={submitPicture} isDisabled={accountBusy}>
          {pictureState.kind === "changing" ? "Updating picture…" : "Change profile picture"}
        </Button>
        {pictureState.kind === "failed" ? <Notice tone="danger">{pictureState.message}</Notice> : null}
        <FieldInput
          label="Username"
          value={username}
          onChangeText={setUsernameDraft}
          autoCapitalize="none"
          editable={!accountBusy}
          placeholder="e.g. river_song"
          hint="3–20 characters: letters, numbers, underscore, dot."
        />
        <Button onPress={submitUsername} isDisabled={accountBusy || !canSubmitUsername}>
          {isBusy ? "Saving…" : "Save username"}
        </Button>
      </Section>

      {user && !user.hasPassword ? (
        <Section label="Password">
          <Typography.Paragraph color="muted">
            You signed up with a social account and have no password yet. Add one to also sign
            in with email.
          </Typography.Paragraph>
          <FieldInput
            label="New password"
            value={newPassword}
            onChangeText={setNewPassword}
            secureTextEntry
            editable={!accountBusy}
            hint="8-15 chars, upper, lower, number & symbol"
          />
          <Button variant="secondary" onPress={submitPassword} isDisabled={accountBusy || !canSubmitPassword}>
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
        <Button variant="outline" onPress={() => void signOutEverywhere()} isDisabled={accountBusy}>
          Sign out everywhere
        </Button>
      </Section>

      <Section label="Notifications">
        <ControlField isSelected={isNotificationsEnabled} onSelectedChange={toggleNotifications}>
          <View className="flex-1">
            <Typography weight="bold">Notifications</Typography>
            <Typography.Paragraph type="body-sm" color="muted">
              Get notified when someone messages you.
            </Typography.Paragraph>
          </View>
          <ControlField.Indicator />
        </ControlField>
      </Section>

      <Section label="Session">
        <Button variant="danger-soft" onPress={() => void signOut()} isDisabled={accountBusy}>
          Sign out
        </Button>
      </Section>

      {saved ? <Notice tone="success">{saved}</Notice> : null}
      {error ? <Notice tone="danger">{error}</Notice> : null}
    </Screen>
  );
}
