import DeviceSizeProvider from "@/components/providers/device-size-provider";
import DeviceTypeProvider from "@/components/providers/device-type-provider";
import NowProvider from "@/components/providers/now-provider";
import { ThemeProvider } from "@/components/providers/theme-provider";
import { Toaster } from "@/components/ui/toast";
import { TooltipProvider } from "@/components/ui/tooltip";
import useKeyboardInsetHeight from "@/lib/hooks/use-keyboard-inset-height";
import React from "react";

export default function Providers({ children }: Readonly<{ children: React.ReactNode }>) {
  useKeyboardInsetHeight();
  return (
    <ThemeProvider>
      <DeviceTypeProvider>
        <DeviceSizeProvider>
          <NowProvider>
            <TooltipProvider>
              {children}
              <Toaster />
            </TooltipProvider>
          </NowProvider>
        </DeviceSizeProvider>
      </DeviceTypeProvider>
    </ThemeProvider>
  );
}
