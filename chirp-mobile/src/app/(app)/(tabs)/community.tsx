import { Typography } from "heroui-native/text";

import { Masthead } from "@/components/masthead";
import { Screen } from "@/components/screen";
import { useTabBarSpace } from "@/components/tab-bar";

export default function Community() {
  const barSpace = useTabBarSpace();

  return (
    <Screen bottomSpace={barSpace}>
      <Masthead title="Community." />
      <Typography.Paragraph color="muted">No communities yet.</Typography.Paragraph>
    </Screen>
  );
}
