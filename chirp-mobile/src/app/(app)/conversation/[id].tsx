import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useLocalSearchParams } from "expo-router";
import { useQuery } from "@tanstack/react-query";
import { FlashList } from "@shopify/flash-list";
import { StatusBar } from "expo-status-bar";
import { Typography } from "heroui-native/text";
import {
  Keyboard,
  KeyboardAvoidingView,
  Platform,
  View,
  type NativeScrollEvent,
  type NativeSyntheticEvent,
} from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { Masthead } from "@/components/masthead";
import { COMPOSER_HEIGHT, MessageInput } from "@/components/message-input";
import { Notice } from "@/components/notice";
import { SafeAreaView } from "@/components/safe-area-view";
import { listContacts } from "@/lib/contacts";
import { goBack } from "@/lib/navigation";
import { spacing } from "@/theme/spacing";
import { bandLabel } from "@/theme/typography";

type ChatMessage = {
  id: string;
  author: "me" | "contact";
  body: string;
  sentAt: number;
};

type ChatRow = ChatMessage & { showTime: boolean };

// Nothing feeds the thread yet, so it starts empty.
const messages: ChatMessage[] = [];

/*
 * The folded composer starts once the thread is scrolled past roughly one
 * and a half bubbles, and unfolds within a finger's travel of the newest
 * message. The gap between the two is hysteresis so it does not chatter
 * while the list settles at the boundary.
 */
const COMPACT_AT = 72;
const EXPAND_AT = 16;

export default function Conversation() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const contactsQuery = useQuery({ queryKey: ["contacts"], queryFn: listContacts });
  const contact = contactsQuery.data?.find((candidate) => candidate.id === id);
  const [composerCompact, setComposerCompact] = useState(false);
  const [keyboardUp, setKeyboardUp] = useState(false);
  const compactRef = useRef(false);
  const insets = useSafeAreaInsets();
  // The keyboard covers the home-indicator inset, so reserving it too would
  // leave the plate floating 34pt above the keys.
  const composerReserve = (keyboardUp ? 0 : insets.bottom) + spacing.sm;

  useEffect(() => {
    const ios = Platform.OS === "ios";
    const shown = Keyboard.addListener(ios ? "keyboardWillShow" : "keyboardDidShow", () =>
      setKeyboardUp(true),
    );
    const hidden = Keyboard.addListener(ios ? "keyboardWillHide" : "keyboardDidHide", () =>
      setKeyboardUp(false),
    );

    return () => {
      shown.remove();
      hidden.remove();
    };
  }, []);

  const rows = useMemo(
    () =>
      messages
        .map((message, index) => ({
          ...message,
          showTime: index === 0 || message.sentAt - messages[index - 1].sentAt >= 12 * 60_000,
        }))
        .reverse(),
    [],
  );

  const handleScroll = useCallback((event: NativeSyntheticEvent<NativeScrollEvent>) => {
    const offset = event.nativeEvent.contentOffset.y;
    const next = offset > COMPACT_AT ? true : offset < EXPAND_AT ? false : compactRef.current;
    if (next === compactRef.current) return;
    compactRef.current = next;
    setComposerCompact(next);
  }, []);

  return (
    <View className="flex-1 bg-background">
      <StatusBar style="dark" />
      {/*
       * The keyboard avoider stays at the screen root so its frame starts at
       * y=0 and the keyboard height it measures is exact. Nesting it inside
       * the safe-area padding makes it undercount by the status-bar inset.
       */}
      <KeyboardAvoidingView className="flex-1" behavior="padding">
        <SafeAreaView edges={["top"]} className="flex-1">
          <View className="flex-1">
            {/* Reserving the composer's footprint puts the list's viewport bottom
                on the composer's bottom edge, so the thread clips there. */}
            <View className="flex-1 gap-6 px-6 pt-4" style={{ paddingBottom: composerReserve }}>
              <Masthead
                meta={contact ? `@${contact.username}` : undefined}
                onBack={() => goBack("/contacts")}
              />
              <View className="flex-1">
                {contactsQuery.isPending ? (
                  <Typography.Paragraph color="muted">Loading conversation…</Typography.Paragraph>
                ) : null}
                {contactsQuery.isError ? (
                  <Notice tone="danger">
                    {contactsQuery.error instanceof Error
                      ? contactsQuery.error.message
                      : "Could not load conversation."}
                  </Notice>
                ) : null}
                {contactsQuery.isSuccess && !contact ? (
                  <Notice tone="danger">Contact not found.</Notice>
                ) : null}
                {contact && rows.length === 0 ? (
                  <Typography.Paragraph color="muted" className="text-center">
                    No messages yet.
                  </Typography.Paragraph>
                ) : null}
                {contact && rows.length > 0 ? (
                  <FlashList
                    data={rows}
                    inverted
                    keyExtractor={(item) => item.id}
                    renderItem={({ item }) => <MessageBubble message={item} />}
                    contentContainerStyle={{
                      paddingTop: COMPOSER_HEIGHT + spacing.md,
                      paddingBottom: spacing.md,
                    }}
                    keyboardShouldPersistTaps="handled"
                    keyboardDismissMode="on-drag"
                    onScroll={handleScroll}
                    scrollEventThrottle={16}
                  />
                ) : null}
              </View>
            </View>
            <View className="absolute left-0 right-0 px-6" style={{ bottom: composerReserve }}>
              <MessageInput compact={composerCompact} />
            </View>
          </View>
        </SafeAreaView>
      </KeyboardAvoidingView>
    </View>
  );
}

function MessageBubble({ message }: { message: ChatRow }) {
  const mine = message.author === "me";

  return (
    <View className="mb-2" style={{ alignItems: mine ? "flex-end" : "flex-start" }}>
      {message.showTime ? (
        <Typography type="body-xs" style={bandLabel} className="mb-1 text-muted">
          {new Date(message.sentAt).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
        </Typography>
      ) : null}
      <View
        accessibilityRole="text"
        className={`px-4 py-2 ${
          mine
            ? "rounded-3xl rounded-br-md bg-accent"
            : "rounded-3xl rounded-bl-md border border-separator bg-surface"
        }`}
        style={{ maxWidth: "78%", borderCurve: "continuous" }}
      >
        <Typography type="body-sm" className={mine ? "text-accent-foreground" : "text-foreground"}>
          {message.body}
        </Typography>
      </View>
    </View>
  );
}
