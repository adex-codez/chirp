import { Link, useRouter } from "expo-router";
import { Button } from "heroui-native/button";
import { ControlField } from "heroui-native/control-field";
import { ListGroup } from "heroui-native/list-group";
import { Separator } from "heroui-native/separator";
import { Typography } from "heroui-native/text";
import { View } from "react-native";

import { BandChip } from "@/components/band-chip";
import { Masthead } from "@/components/masthead";
import { Notice } from "@/components/notice";
import { Screen } from "@/components/screen";
import { Section } from "@/components/section";
import { Songline } from "@/components/songline";
import { useAppStore } from "@/store/use-app-store";
import { useSessionStore } from "@/store/use-session-store";

export default function Index() {
  const router = useRouter();
  const isNotificationsEnabled = useAppStore((state) => state.isNotificationsEnabled);
  const toggleNotifications = useAppStore((state) => state.toggleNotifications);
  const sessionUser = useSessionStore((state) => state.user);
  const error = useSessionStore((state) => state.error);
  const isBusy = useSessionStore((state) => state.isBusy);
  const signOut = useSessionStore((state) => state.signOut);

  // `(app)` sits behind the authenticated guard, so a null User means the
  // session was dropped while this screen was mounted.
  if (!sessionUser) {
    return (
      <Screen>
        <Masthead title="Signed out." />
        <Notice tone="muted">You are not signed in. Sign in to use Chirp.</Notice>
        <Link href="/sign-in" asChild>
          <Button>Sign in</Button>
        </Link>
      </Screen>
    );
  }

  return (
    <Screen>
      <Masthead title="Your account is ready." meta="SIGNED IN">
        <BandChip username={sessionUser.username} verified={sessionUser.verified} />
        <Songline height={28} />
      </Masthead>

      <Section label="Account">
        <ListGroup variant="transparent" className="-mx-4 -my-4">
          <ListGroup.Item className="px-0">
            <ListGroup.ItemContent>
              <ListGroup.ItemTitle>Email</ListGroup.ItemTitle>
              <ListGroup.ItemDescription>{sessionUser.email}</ListGroup.ItemDescription>
            </ListGroup.ItemContent>
          </ListGroup.Item>
          <Separator />
          <ListGroup.Item className="px-0" onPress={() => router.push("/settings")}>
            <ListGroup.ItemContent>
              <ListGroup.ItemTitle>Settings</ListGroup.ItemTitle>
              <ListGroup.ItemDescription>
                Username, password, devices
              </ListGroup.ItemDescription>
            </ListGroup.ItemContent>
            <ListGroup.ItemSuffix />
          </ListGroup.Item>
        </ListGroup>
      </Section>

      <Section label="Notifications">
        <ControlField isSelected={isNotificationsEnabled} onSelectedChange={toggleNotifications}>
          <View className="flex-1">
            <Typography weight="bold">Notifications</Typography>
            <Typography.Paragraph type="body-sm" color="muted">
              Get notified when someone messages you.
            </Typography.Paragraph>
          </View>
          <ControlField.Indicator />
        </ControlField>
      </Section>

      {error ? <Notice tone="danger">{error}</Notice> : null}

      <Section label="Session">
        <Button variant="danger-soft" onPress={() => void signOut()} isDisabled={isBusy}>
          Sign out
        </Button>
      </Section>
    </Screen>
  );
}
