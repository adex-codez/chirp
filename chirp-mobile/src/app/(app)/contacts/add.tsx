import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { Button } from "@/components/button";
import { FieldInput } from "@/components/field-input";
import { Masthead } from "@/components/masthead";
import { Notice } from "@/components/notice";
import { Screen } from "@/components/screen";
import { addContact } from "@/lib/contacts";
import { goBack } from "@/lib/navigation";

export default function AddContact() {
  const [username, setUsername] = useState("");
  const queryClient = useQueryClient();

  const addContactMutation = useMutation({
    mutationFn: addContact,
    onSuccess: async () => {
      setUsername("");
      await queryClient.invalidateQueries({ queryKey: ["contacts"] });
      goBack("/contacts");
    },
  });

  const canAddContact = username.trim().length > 0;

  const submitAddContact = () => {
    if (addContactMutation.isPending || !canAddContact) return;
    addContactMutation.mutate(username.trim());
  };

  return (
    <Screen>
      <Masthead title="Add a contact." onBack={() => goBack("/contacts")} />
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
        isDisabled={addContactMutation.isPending || !canAddContact}
        onPress={submitAddContact}
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
    </Screen>
  );
}
