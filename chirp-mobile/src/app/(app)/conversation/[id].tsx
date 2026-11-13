import { useLocalSearchParams } from "expo-router";
import { useQuery } from "@tanstack/react-query";
import { Typography } from "heroui-native/text";

import { Masthead } from "@/components/masthead";
import { Notice } from "@/components/notice";
import { Screen } from "@/components/screen";
import { listContacts } from "@/lib/contacts";
import { goBack } from "@/lib/navigation";

export default function Conversation() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const contactsQuery = useQuery({ queryKey: ["contacts"], queryFn: listContacts });
  const contact = contactsQuery.data?.find((candidate) => candidate.id === id);

  return (
    <Screen>
      <Masthead
        title="Conversation."
        meta={contact ? `@${contact.username}` : undefined}
        onBack={() => goBack("/contacts")}
      />
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
      {contact ? (
        <Typography.Paragraph color="muted">No messages yet.</Typography.Paragraph>
      ) : null}
    </Screen>
  );
}
