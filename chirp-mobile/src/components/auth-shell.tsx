import type { ReactNode } from "react";
import { KeyboardAvoidingView, Platform, ScrollView, View } from "react-native";
import { StatusBar } from "expo-status-bar";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { Typography } from "heroui-native/text";

import { ChirpLogo } from "@/components/chirp-logo";
import { SafeAreaView } from "@/components/safe-area-view";
import { Songline } from "@/components/songline";
import { spacing } from "@/theme/spacing";
import { bandLabel, displayTitle } from "@/theme/typography";

type AuthShellProps = {
  kicker: string;
  title: ReactNode;
  intro?: string;
  step?: number;
  of?: number;
  children: ReactNode;
};

export function AuthShell({ kicker, title, intro, step, of = 3, children }: AuthShellProps) {
  const insets = useSafeAreaInsets();

  return (
    <View className="flex-1 bg-background">
      <StatusBar style="light" />
      <SafeAreaView edges={["top"]} className="bg-night">
        <View className="bg-night px-6 pb-9 pt-3">
          <View className="flex-row items-center justify-between">
            <ChirpLogo size={32} />
            <Typography style={bandLabel} type="body-xs" className="text-dawn">
              {kicker}
            </Typography>
          </View>
          <Typography.Heading style={displayTitle} type="h1" className="mt-6 text-white">
            {title}
          </Typography.Heading>
          {intro ? (
            <Typography.Paragraph className="mt-2 text-brand-200">{intro}</Typography.Paragraph>
          ) : null}
          <View className="mt-6 flex-row items-center justify-between">
            <Songline active={step === undefined ? undefined : Math.round((step / of) * 20)} />
            {step === undefined ? null : (
              <Typography style={bandLabel} type="body-xs" className="text-brand-200">
                {String(step).padStart(2, "0")} / {String(of).padStart(2, "0")}
              </Typography>
            )}
          </View>
        </View>
      </SafeAreaView>
      <KeyboardAvoidingView
        className="flex-1"
        behavior={Platform.OS === "ios" ? "padding" : undefined}
      >
        <ScrollView
          className="-mt-6 flex-1 rounded-t-3xl bg-background"
          style={{ borderCurve: "continuous" }}
          contentContainerClassName="grow gap-5 px-6 pt-7"
          contentContainerStyle={{ paddingBottom: insets.bottom + spacing.lg }}
          keyboardShouldPersistTaps="handled"
          keyboardDismissMode="on-drag"
        >
          {children}
        </ScrollView>
      </KeyboardAvoidingView>
    </View>
  );
}
