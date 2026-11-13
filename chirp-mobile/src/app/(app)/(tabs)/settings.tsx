import { useThemeColor } from "heroui-native/hooks";
import { ControlField } from "heroui-native/control-field";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { Pressable, View } from "react-native";

import { Button } from "@/components/button";
import { BandChip } from "@/components/band-chip";
import { Camera, ChevronLeft } from "@/components/icons";
import { FieldInput } from "@/components/field-input";
import { Masthead } from "@/components/masthead";
import { Notice } from "@/components/notice";
import { ProfileAvatar } from "@/components/profile-avatar";
import { Screen } from "@/components/screen";
import { Section } from "@/components/section";
import { useTabBarSpace } from "@/components/tab-bar";
import { spacing } from "@/theme";
import { useAppStore } from "@/store/use-app-store";
import { useSessionStore } from "@/store/use-session-store";

type OpenEditor = "username" | "password" | null;

export default function Settings() {
  const barSpace = useTabBarSpace();
  const [accentForeground, muted] = useThemeColor(["accent-foreground", "muted"]);
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
  const [openEditor, setOpenEditor] = useState<OpenEditor>(null);
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
      setOpenEditor(null);
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
      setOpenEditor(null);
      setSaved("Password added. Both sign-in methods reach this User now.");
    } catch {
      // Error text already lives in the store.
    }
  };

  return (
    <Screen bottomSpace={barSpace}>
      <Masthead title="Settings." />

      {user ? (
        <View className="items-center gap-3 py-2">
          <View className="relative">
            <ProfileAvatar subject={user} size="lg" />
            <Button
              variant="primary"
              size="sm"
              isIconOnly
              onPress={submitPicture}
              isDisabled={accountBusy}
              accessibilityLabel="Upload profile picture"
              hitSlop={10}
              className="absolute rounded-full border-2 border-background"
              style={{
                position: "absolute",
                right: -spacing.xs,
                bottom: -spacing.xs,
                width: spacing.md + spacing.sm,
                height: spacing.md + spacing.sm,
              }}
            >
              <Camera size={14} color={accentForeground} />
            </Button>
          </View>
          <Typography.Heading type="h5">{user.username}</Typography.Heading>
          {pictureState.kind === "changing" ? (
            <Typography.Paragraph type="body-sm" color="muted">
              Updating profile picture…
            </Typography.Paragraph>
          ) : null}
          {pictureState.kind === "failed" ? <Notice tone="danger">{pictureState.message}</Notice> : null}
        </View>
      ) : null}

      <Section label="Account">
        <Pressable
          accessibilityRole="button"
          accessibilityState={{ expanded: openEditor === "username" }}
          onPress={() => setOpenEditor(openEditor === "username" ? null : "username")}
          disabled={accountBusy}
          className="min-h-14 flex-row items-center gap-4 border-b border-separator px-1"
          style={({ pressed }) => ({ opacity: pressed ? 0.65 : 1 })}
        >
          <View className="flex-1">
            <Typography>Username</Typography>
            <Typography.Paragraph type="body-sm" color="muted">
              @{user?.username ?? username}
            </Typography.Paragraph>
          </View>
          <View style={{ transform: [{ rotate: "180deg" }] }}>
            <ChevronLeft color={muted} />
          </View>
        </Pressable>
        {openEditor === "username" ? (
          <View className="gap-3 pt-3">
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
          </View>
        ) : null}

        {user && !user.hasPassword ? (
          <>
            <Pressable
              accessibilityRole="button"
              accessibilityState={{ expanded: openEditor === "password" }}
              onPress={() => setOpenEditor(openEditor === "password" ? null : "password")}
              disabled={accountBusy}
              className="min-h-14 flex-row items-center gap-4 border-b border-separator px-1"
              style={({ pressed }) => ({ opacity: pressed ? 0.65 : 1 })}
            >
              <Typography className="w-6 text-center text-muted">⌑</Typography>
              <View className="flex-1">
                <Typography>Password</Typography>
                <Typography.Paragraph type="body-sm" color="muted">
                  Add a password for email sign-in
                </Typography.Paragraph>
              </View>
              <View style={{ transform: [{ rotate: "180deg" }] }}>
                <ChevronLeft color={muted} />
              </View>
            </Pressable>
            {openEditor === "password" ? (
              <View className="gap-3 pt-3">
                <FieldInput
                  label="New password"
                  value={newPassword}
                  onChangeText={setNewPassword}
                  secureTextEntry
                  editable={!accountBusy}
                  hint="8-15 chars, upper, lower, number & symbol"
                />
                <Button
                  variant="secondary"
                  onPress={submitPassword}
                  isDisabled={accountBusy || !canSubmitPassword}
                >
                  {isBusy ? "Saving…" : "Add password"}
                </Button>
              </View>
            ) : null}
          </>
        ) : null}
      </Section>

      <Section label="Preferences">
        <ControlField isSelected={isNotificationsEnabled} onSelectedChange={toggleNotifications}>
          <View className="flex-1">
            <Typography>Notifications</Typography>
            <Typography.Paragraph type="body-sm" color="muted">
              Get notified when someone messages you.
            </Typography.Paragraph>
          </View>
          <ControlField.Indicator />
        </ControlField>
      </Section>

      <Section label="Devices">
        <Typography.Paragraph color="muted">
          {user
            ? `Signed in as ${user.email} on this device.`
            : "Signing out everywhere ends every other session."}
        </Typography.Paragraph>
        <Button variant="outline" onPress={() => void signOutEverywhere()} isDisabled={accountBusy}>
          Sign out everywhere
        </Button>
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
