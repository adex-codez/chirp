import { create } from "zustand";

type AppStore = {
  isNotificationsEnabled: boolean;
  toggleNotifications: () => void;
};

export const useAppStore = create<AppStore>((set) => ({
  isNotificationsEnabled: true,
  toggleNotifications: () =>
    set((state) => ({
      isNotificationsEnabled: !state.isNotificationsEnabled,
    })),
}));
