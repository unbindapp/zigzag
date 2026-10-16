import { Switch as SwitchPrimitive } from "@base-ui/react/switch";

import { ToggleKnob } from "@/components/ui/toggle-knob";
import { cn } from "@/components/ui/utils";

type TSwitchProps = Omit<SwitchPrimitive.Root.Props, "children"> & {
  size?: "sm" | "md";
};

function Switch({ className, checked, size = "sm", ...props }: TSwitchProps) {
  return (
    <SwitchPrimitive.Root
      data-slot="switch"
      checked={checked}
      className={cn(
        "group/switch focus-visible:ring-primary/8-10 focus-visible:ring-offset-background inline-flex shrink-0 cursor-pointer items-center rounded-full outline-hidden focus-visible:ring-1 focus-visible:ring-offset-2 data-disabled:cursor-not-allowed data-disabled:opacity-50",
        className,
      )}
      {...props}
    >
      <SwitchKnob size={size} />
    </SwitchPrimitive.Root>
  );
}

function SwitchKnob({ size }: { size: "sm" | "md" }) {
  return (
    <SwitchPrimitive.Thumb
      render={(thumbProps, state) => (
        <span {...thumbProps}>
          <ToggleKnob checked={state.checked} size={size} />
        </span>
      )}
    />
  );
}

export { Switch };
