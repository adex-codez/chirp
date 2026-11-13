import { Pressable, View } from "react-native";
import { router } from "expo-router";
import { useQuery } from "@tanstack/react-query";
import { useThemeColor } from "heroui-native/hooks";
import { Typography } from "heroui-native/text";

import { Plus } from "@/components/icons";
import { Masthead } from "@/components/masthead";
import { Notice } from "@/components/notice";
import { ProfileAvatar } from "@/components/profile-avatar";
import { Screen } from "@/components/screen";
import { listContacts } from "@/lib/contacts";
import { goBack } from "@/lib/navigation";
import { spacing } from "@/theme/spacing";

export default function Contacts() {
  const accent = useThemeColor("accent");
  const contactsQuery = useQuery({ queryKey: ["contacts"], queryFn: listContacts });

  return (
    <Screen>
      <Masthead title="Start a conversation." onBack={() => goBack("/")}>
        <Typography.Paragraph color="muted">
          Choose a contact or add someone new.
        </Typography.Paragraph>
      </Masthead>

      <Pressable
        accessibilityRole="button"
        onPress={() => router.push("/contacts/add")}
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
                onPress={() =>
                  router.push({
                    pathname: "/conversation/[id]",
                    params: { id: contact.id },
                  })
                }
                style={({ pressed }) => [pressed ? { opacity: 0.6 } : null]}
              >
                <View className="flex flex-row items-center gap-2">
                  <ProfileAvatar subject={contact} />
                  <Typography weight="bold">@{contact.username}</Typography>
                </View>
              </Pressable>
            ))
          : null}
      </View>
    </Screen>
  );
}
