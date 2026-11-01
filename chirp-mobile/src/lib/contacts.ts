import { apiAuth } from "@/lib/api";

export type Contact = {
  id: string;
  username: string;
};

export function listContacts(): Promise<Contact[]> {
  return apiAuth<Contact[]>("/contacts");
}

export function addContact(username: string): Promise<Contact> {
  return apiAuth<Contact>("/contacts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username }),
  });
}
