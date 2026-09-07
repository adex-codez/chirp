import { Button } from "heroui-native/button";
import { Card } from "heroui-native/card";
import { Typography } from "heroui-native/text";
import { ScrollView, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { useAppStore } from "@/store/use-app-store";

export default function Index() {
  const isNotificationsEnabled = useAppStore(
    (state) => state.isNotificationsEnabled,
  );
  const toggleNotifications = useAppStore(
    (state) => state.toggleNotifications,
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
