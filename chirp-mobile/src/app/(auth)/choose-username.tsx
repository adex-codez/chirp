import { Button } from "heroui-native/button";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { ScrollView, View } from "react-native";
import { AuthShell } from "@/components/auth-shell";
import { FieldInput } from "@/components/field-input";

import { useSessionStore } from "@/store/use-session-store";

export default function ChooseUsername() {
  const [draft, setDraft] = useState("");
  const user = useSessionStore((state) => state.user);
  const error = useSessionStore((state) => state.error);
  const isBusy = useSessionStore((state) => state.isBusy);
  const setUsername = useSessionStore((state) => state.setUsername);
  const cancelUsernamePick = useSessionStore((state) => state.cancelUsernamePick);

  const submit = async () => {
    try {
      await setUsername(draft.trim());
    } catch {
      // Error text (taken handle, bad shape) already lives in the store.
    }
  };

  return (
    <AuthShell
      kicker="ONE LAST THING"
      title="Pick your call sign."
      intro={
        user
          ? `You signed in with Social sign-in as ${user.email}. Now claim the unique, public Username the chorus will know you by — nothing else unlocks until it is set.`
          : "You signed in with Social sign-in. Now claim your unique, public Username — nothing else unlocks until it is set."
      }
      step={3}
    >
      <ScrollView className="flex-1" contentContainerClassName="grow gap-5 pb-8">
        <FieldInput
          label="Username"
          value={draft}
          onChangeText={setDraft}
          autoCapitalize="none"
          editable={!isBusy}
          placeholder="e.g. river_song"
          hint="3–20 characters: letters, numbers, underscore, dot."
        />
        {error ? (
          <Typography.Paragraph className="text-danger">{error}</Typography.Paragraph>
        ) : null}
        <View className="gap-3">
          <Button onPress={submit} isDisabled={isBusy}>
            {isBusy ? "Saving…" : "Save Username"}
          </Button>
          <Button variant="outline" onPress={() => void cancelUsernamePick()} isDisabled={isBusy}>
            Cancel and sign out
          </Button>
        </View>
      </ScrollView>
    </AuthShell>
  );
}
