import { Typography } from "heroui-native/text";

import { Masthead } from "@/components/masthead";
import { Screen } from "@/components/screen";
import { useTabBarSpace } from "@/components/tab-bar";

export default function Call() {
  const barSpace = useTabBarSpace();

  return (
    <Screen bottomSpace={barSpace}>
      <Masthead title="Call." />
      <Typography.Paragraph color="muted">No calls yet.</Typography.Paragraph>
    </Screen>
  );
}
