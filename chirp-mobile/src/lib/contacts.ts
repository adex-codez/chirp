import { apiAuth } from "@/lib/api";

type ContactPayload = {
  id: string;
  username: string;
  profile_picture_url?: string | null;
};

export type Contact = {
  id: string;
  username: string;
  profilePictureUrl: string | null;
};

function toContact(value: ContactPayload): Contact {
  return {
    id: value.id,
    username: value.username,
    profilePictureUrl: value.profile_picture_url?.trim() || null,
  };
}

export async function listContacts(): Promise<Contact[]> {
  const contacts = await apiAuth<ContactPayload[]>("/contacts");
  return contacts.map(toContact);
}

export async function addContact(username: string): Promise<Contact> {
  const contact = await apiAuth<ContactPayload>("/contacts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username }),
  });
  return toContact(contact);
}
