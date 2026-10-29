import { Tabs } from "expo-router/js-tabs";

import { TAB_DESTINATIONS, TabBar } from "@/components/tab-bar";

export default function AppGroupLayout() {
  return (
    <Tabs
      screenOptions={{ headerShown: false, animation: "none" }}
      tabBar={(props) => <TabBar {...props} />}
    >
      {TAB_DESTINATIONS.map((destination) => (
        <Tabs.Screen key={destination.route} name={destination.route} />
      ))}
    </Tabs>
  );
}
