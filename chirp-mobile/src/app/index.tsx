import { Redirect } from "expo-router";

import { useSessionStore } from "@/store/use-session-store";

// `/` has no content of its own: land here only transiently, then hand
// off to whichever group the session status allows. The root layout
// holds on a debug screen while restoring, so this renders only once a
// real status (guest, pending-*, authenticated) is known.
export default function Index() {
  const status = useSessionStore((state) => state.status);

  if (status === "authenticated") {
    return <Redirect href="/(app)" />;
  }
  if (status === "pending-verification") {
    return <Redirect href="/verify" />;
  }
  if (status === "needs-username") {
    return <Redirect href="/choose-username" />;
  }
  return <Redirect href="/sign-in" />;
}
