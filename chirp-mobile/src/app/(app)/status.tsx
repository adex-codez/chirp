import { Typography } from "heroui-native/text";

import { Masthead } from "@/components/masthead";
import { Screen } from "@/components/screen";
import { useTabBarSpace } from "@/components/tab-bar";

export default function Status() {
  const barSpace = useTabBarSpace();

  return (
    <Screen bottomSpace={barSpace}>
      <Masthead title="Status." />
      <Typography.Paragraph color="muted">No status updates yet.</Typography.Paragraph>
    </Screen>
  );
}
