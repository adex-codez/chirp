import { useState } from "react";
import { Pressable, View } from "react-native";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/button";
import { useThemeColor } from "heroui-native/hooks";
import { Typography } from "heroui-native/text";

import { FieldInput } from "@/components/field-input";
import { Plus } from "@/components/icons";
import { Masthead } from "@/components/masthead";
import { Notice } from "@/components/notice";
import { Screen } from "@/components/screen";
import { useTabBarSpace } from "@/components/tab-bar";
import { addContact, listContacts, type Contact } from "@/lib/contacts";
import { spacing } from "@/theme/spacing";

type ChatScreen =
  | { kind: "inbox" }
  | { kind: "contacts" }
  | { kind: "add-contact" }
  | { kind: "conversation"; contact: Contact };

export default function Chat() {
  const barSpace = useTabBarSpace();
  const [screen, setScreen] = useState<ChatScreen>({ kind: "inbox" });
  const [username, setUsername] = useState("");
  const [accent, accentForeground] = useThemeColor(["accent", "accent-foreground"]);
  const queryClient = useQueryClient();
  const contactsQuery = useQuery({
    queryKey: ["contacts"],
    queryFn: listContacts,
    enabled: screen.kind === "contacts",
  });
  const addContactMutation = useMutation({
    mutationFn: addContact,
    onSuccess: async () => {
      setUsername("");
      await queryClient.invalidateQueries({ queryKey: ["contacts"] });
      setScreen({ kind: "contacts" });
    },
  });

  const goBack = () => {
    setScreen(screen.kind === "contacts" ? { kind: "inbox" } : { kind: "contacts" });
  };

  return (
    <Screen
      bottomSpace={barSpace}
      overlay={
        screen.kind === "inbox" ? (
          <Pressable
            accessibilityRole="button"
            accessibilityLabel="Start a conversation"
            onPress={() => setScreen({ kind: "contacts" })}
            className="absolute right-6 h-14 w-14 items-center justify-center rounded-full bg-accent shadow-overlay"
            style={({ pressed }) => [
              { bottom: barSpace + spacing.md, borderCurve: "continuous" },
              pressed ? { opacity: 0.75 } : null,
            ]}
          >
            <Plus color={accentForeground} />
          </Pressable>
        ) : null
      }
    >
      {screen.kind === "inbox" ? (
        <>
          <Masthead title="Chat." />
          <Typography.Paragraph color="muted">No conversations yet.</Typography.Paragraph>
        </>
      ) : null}

      {screen.kind === "contacts" ? (
        <>
          <Masthead title="Start a conversation." onBack={goBack}>
            <Typography.Paragraph color="muted">
              Choose a contact or add someone new.
            </Typography.Paragraph>
          </Masthead>

          <Pressable
            accessibilityRole="button"
            onPress={() => setScreen({ kind: "add-contact" })}
            style={({ pressed }) => [
              { flexDirection: "row", alignItems: "center", gap: spacing.md },
              pressed ? { opacity: 0.6 } : null,
            ]}
          >
            <View className="h-11 w-11 items-center justify-center rounded-full border border-separator bg-surface">
              <Plus color={accent} size={20} />
            </View>
            <View className="flex-1">
              <Typography weight="bold">Add a contact</Typography>
              <Typography type="body-sm" color="muted">
                Find someone by username
              </Typography>
            </View>
          </Pressable>

          <View style={{ gap: spacing.md, marginTop: spacing.xxl }}>
            {contactsQuery.isPending ? (
              <Typography.Paragraph color="muted">Loading contacts…</Typography.Paragraph>
            ) : null}
            {contactsQuery.isError ? (
              <Notice tone="danger">
                {contactsQuery.error instanceof Error
                  ? contactsQuery.error.message
                  : "Could not load contacts."}
              </Notice>
            ) : null}
            {contactsQuery.isSuccess && contactsQuery.data.length === 0 ? (
              <Typography.Paragraph color="muted">
                No contacts yet. Add someone by username to start a conversation.
              </Typography.Paragraph>
            ) : null}
            {contactsQuery.isSuccess && contactsQuery.data.length > 0
              ? contactsQuery.data.map((contact) => (
                  <Pressable
                    key={contact.id}
                    accessibilityRole="button"
                    accessibilityLabel={`Start a conversation with ${contact.username}`}
                    onPress={() => setScreen({ kind: "conversation", contact })}
                    style={({ pressed }) => [pressed ? { opacity: 0.6 } : null]}
                  >
                    <Typography weight="bold">@{contact.username}</Typography>
                  </Pressable>
                ))
              : null}
          </View>
        </>
      ) : null}

      {screen.kind === "add-contact" ? (
        <>
          <Masthead title="Add a contact." onBack={goBack} />
          <FieldInput
            label="Username"
            value={username}
            onChangeText={setUsername}
            autoCapitalize="none"
            autoCorrect={false}
            placeholder="Enter username"
            editable={!addContactMutation.isPending}
          />
          <Button
            isDisabled={!username.trim() || addContactMutation.isPending}
            onPress={() => addContactMutation.mutate(username.trim())}
          >
            {addContactMutation.isPending ? "Adding contact…" : "Add contact"}
          </Button>
          {addContactMutation.isError ? (
            <Notice tone="danger">
              {addContactMutation.error instanceof Error
                ? addContactMutation.error.message
                : "Could not add contact."}
            </Notice>
          ) : null}
        </>
      ) : null}

      {screen.kind === "conversation" ? (
        <>
          <Masthead title="Conversation." meta={`@${screen.contact.username}`} onBack={goBack} />
          <Typography.Paragraph color="muted">No messages yet.</Typography.Paragraph>
        </>
      ) : null}
    </Screen>
  );
}
