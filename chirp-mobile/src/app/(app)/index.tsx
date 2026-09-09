import { Link } from "expo-router";
import { Button } from "heroui-native/button";
import { Card } from "heroui-native/card";
import { Typography } from "heroui-native/text";
import { ScrollView, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { useAppStore } from "@/store/use-app-store";
import { useSessionStore } from "@/store/use-session-store";

export default function Index() {
  const isNotificationsEnabled = useAppStore(
    (state) => state.isNotificationsEnabled,
  );
  const toggleNotifications = useAppStore(
    (state) => state.toggleNotifications,
  );
  const sessionStatus = useSessionStore((state) => state.status);
  const sessionUser = useSessionStore((state) => state.user);
  const signOut = useSessionStore((state) => state.signOut);
  const signOutEverywhere = useSessionStore(
    (state) => state.signOutEverywhere,
  );

  return (
    <SafeAreaView className="flex-1 bg-background">
      <ScrollView
        className="flex-1"
        contentContainerClassName="grow gap-8 px-6 py-8"
      >
        <View className="gap-3">
          <Typography type="body-xs" weight="bold" className="text-accent">
            CHIRP MOBILE
          </Typography>
          <Typography.Heading type="h1">
            Your social space is ready.
          </Typography.Heading>
          <Typography.Paragraph color="muted">
            HeroUI and Uniwind handle the interface. Zustand owns local state,
            and TanStack Query is ready for server data.
          </Typography.Paragraph>
        </View>

        <Card variant="secondary">
          <Card.Body className="gap-3">
            <Typography.Heading type="h3">Session</Typography.Heading>
            {sessionStatus === "authenticated" && sessionUser ? (
              <Typography.Paragraph color="muted">
                Signed in as {sessionUser.username} ({sessionUser.email}).
              </Typography.Paragraph>
            ) : sessionStatus === "pending-verification" ? (
              <Typography.Paragraph color="muted">
                Your Email is not verified yet — enter the code to continue.
              </Typography.Paragraph>
            ) : (
              <Typography.Paragraph color="muted">
                You are not signed in yet. Sign in to use Chirp.
              </Typography.Paragraph>
            )}
          </Card.Body>
          <Card.Footer className="flex-col items-stretch gap-3">
            {sessionStatus === "authenticated" ? (
              <>
                <Button variant="outline" onPress={() => void signOut()}>
                  Sign out
                </Button>
                <Button variant="outline" onPress={() => void signOutEverywhere()}>
                  Sign out everywhere
                </Button>
              </>
            ) : sessionStatus === "pending-verification" ? (
              <Link href="/verify" asChild>
                <Button>Enter verification code</Button>
              </Link>
            ) : (
              <>
                <Link href="/sign-in" asChild>
                  <Button>Sign in</Button>
                </Link>
                <Link href="/sign-up" asChild>
                  <Button variant="outline">Sign up</Button>
                </Link>
              </>
            )}
          </Card.Footer>
        </Card>

        <Card variant="secondary">
          <Card.Body className="gap-3">
            <Typography.Heading type="h3">App foundation</Typography.Heading>
            <Typography.Paragraph color="muted">
              This button reads and updates a Zustand store, so the shared app
              state is wired from the first screen.
            </Typography.Paragraph>
          </Card.Body>
          <Card.Footer>
            <Button
              variant={isNotificationsEnabled ? "primary" : "outline"}
              onPress={toggleNotifications}
            >
              {isNotificationsEnabled
                ? "Notifications on"
                : "Notifications off"}
            </Button>
          </Card.Footer>
        </Card>
      </ScrollView>
    </SafeAreaView>
  );
}
