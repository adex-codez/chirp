import { Button } from "heroui-native/button";
import { Card } from "heroui-native/card";
import { Typography } from "heroui-native/text";
import { useState } from "react";
import { ScrollView, TextInput, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { useSessionStore } from "@/store/use-session-store";

export default function ChooseUsername() {
  const [draft, setDraft] = useState("");
  const user = useSessionStore((state) => state.user);
  const error = useSessionStore((state) => state.error);
  const isBusy = useSessionStore((state) => state.isBusy);
  const setUsername = useSessionStore((state) => state.setUsername);
  const cancelUsernamePick = useSessionStore(
    (state) => state.cancelUsernamePick,
  );

  const submit = async () => {
    try {
      await setUsername(draft.trim());
    } catch {
      // Error text (taken handle, bad shape) already lives in the store.
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
            ALMOST THERE
          </Typography>
          <Typography.Heading type="h1">Pick your Username.</Typography.Heading>
          <Typography.Paragraph color="muted">
            {user
              ? `Signed in as ${user.email}. Your Username is unique and public — nothing else unlocks until it is set.`
              : "Your Username is unique and public — nothing else unlocks until it is set."}
          </Typography.Paragraph>
        </View>

        <Card variant="secondary">
          <Card.Body className="gap-4">
            <View className="gap-1">
              <Typography type="body-sm" weight="bold">
                Username
              </Typography>
              <TextInput
                value={draft}
                onChangeText={setDraft}
                autoCapitalize="none"
                placeholder="e.g. river_song"
                className="rounded-lg border border-separator px-3 py-2 text-foreground"
              />
              <Typography.Paragraph color="muted">
                3–20 characters: letters, numbers, underscore, dot.
              </Typography.Paragraph>
            </View>
            {error ? (
              <Typography.Paragraph className="text-danger">
                {error}
              </Typography.Paragraph>
            ) : null}
          </Card.Body>
          <Card.Footer className="flex-col items-stretch gap-3">
            <Button onPress={submit} isDisabled={isBusy}>
              {isBusy ? "Saving…" : "Save Username"}
            </Button>
            <Button
              variant="outline"
              onPress={() => void cancelUsernamePick()}
              isDisabled={isBusy}
            >
              Cancel and sign out
            </Button>
          </Card.Footer>
        </Card>
      </ScrollView>
    </SafeAreaView>
  );
}
