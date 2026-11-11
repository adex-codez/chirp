export type SessionUser = {
  id: string;
  username: string;
  email: string;
  verified: boolean;
  hasPassword: boolean;
  profilePictureUrl: string | null;
};

function parseUser(value: unknown, source: "http" | "storage"): SessionUser {
  if (
    typeof value !== "object" ||
    value === null ||
    !("id" in value) ||
    typeof value.id !== "string" ||
    value.id.length === 0 ||
    !("username" in value) ||
    typeof value.username !== "string" ||
    !("email" in value) ||
    typeof value.email !== "string" ||
    !("verified" in value) ||
    typeof value.verified !== "boolean"
  ) {
    throw new Error("Your User details could not be read. Try again.");
  }

  const hasPassword =
    source === "http"
      ? "has_password" in value
        ? value.has_password
        : undefined
      : "hasPassword" in value
        ? value.hasPassword ?? true
        : true;
  const picture =
    source === "http"
      ? "profile_picture_url" in value
        ? value.profile_picture_url
        : null
      : "profilePictureUrl" in value
        ? value.profilePictureUrl
        : null;

  if (
    typeof hasPassword !== "boolean" ||
    (picture !== null && picture !== undefined && typeof picture !== "string")
  ) {
    throw new Error("Your User details could not be read. Try again.");
  }

  return {
    id: value.id,
    username: value.username,
    email: value.email,
    verified: value.verified,
    hasPassword,
    profilePictureUrl: picture || null,
  };
}

export function parseSessionUser(value: unknown): SessionUser {
  return parseUser(value, "http");
}

export function parsePersistedSessionUser(value: unknown): SessionUser {
  return parseUser(value, "storage");
}
