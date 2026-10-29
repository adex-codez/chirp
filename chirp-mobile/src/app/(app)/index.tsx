import { Typography } from "heroui-native/text";

import { Masthead } from "@/components/masthead";
import { Screen } from "@/components/screen";
import { useTabBarSpace } from "@/components/tab-bar";

export default function Chat() {
  const barSpace = useTabBarSpace();

  return (
    <Screen bottomSpace={barSpace}>
      <Masthead title="Chat." />
      <Typography.Paragraph color="muted">No conversations yet.</Typography.Paragraph>
    </Screen>
  );
}
