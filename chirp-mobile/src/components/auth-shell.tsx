import type { ReactNode } from "react";
import { View } from "react-native";
import { Typography } from "heroui-native/text";
import { ChirpLogo } from "@/components/chirp-logo";
import { Songline } from "@/components/songline";
import { SafeAreaView } from "@/components/safe-area-view";
import { fontStack } from "@/theme/typography";

/*
 * AuthShell — night band header over a paper form.
 * The header is the thesis: a banding station at first light.
 * `step` / `of` encodes real progress through the verification
 * sequence (claim → code → call sign), so the meter is information.
 */

type AuthShellProps = {
  kicker: string;
  title: ReactNode;
  intro?: string;
  step?: number;
  of?: number;
  children: ReactNode;
};

export function AuthShell({ kicker, title, intro, step, of = 3, children }: AuthShellProps) {
  return (
    <SafeAreaView className="flex-1 bg-background">
      <View className="bg-night px-6 pb-7 pt-10">
        <View className="flex-row items-center justify-between">
          <ChirpLogo size={40} monochrome />
          <Typography
            style={{ fontFamily: fontStack.band, letterSpacing: 2 }}
            type="body-xs"
            className="text-dawn"
          >
            {kicker}
          </Typography>
        </View>
        <Typography
          style={{ fontFamily: fontStack.display, fontStyle: "italic" }}
          type="h1"
          className="mt-5 text-white"
        >
          {title}
        </Typography>
        {intro ? (
          <Typography.Paragraph className="mt-2 text-brand-200">
            {intro}
          </Typography.Paragraph>
        ) : null}
        <View className="mt-5 flex-row items-center justify-between">
          <Songline active={step === undefined ? undefined : Math.round((step / of) * 20)} />
          {step !== undefined ? (
            <Typography
              style={{ fontFamily: fontStack.band, letterSpacing: 1.5 }}
              type="body-xs"
              className="text-brand-200"
            >
              {String(step).padStart(2, "0")} / {String(of).padStart(2, "0")}
            </Typography>
          ) : null}
        </View>
      </View>
      <View className="flex-1 px-6 py-6">{children}</View>
    </SafeAreaView>
  );
}
