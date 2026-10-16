import { Tabs as TabsPrimitive } from "@base-ui/react/tabs";

import { cn } from "@/components/ui/utils";

function Tabs({ className, ...props }: TabsPrimitive.Root.Props) {
  return (
    <TabsPrimitive.Root
      data-slot="tabs"
      className={cn("flex flex-col gap-4", className)}
      {...props}
    />
  );
}

function TabsList({ className, children, ...props }: TabsPrimitive.List.Props) {
  return (
    <TabsPrimitive.List
      data-slot="tabs-list"
      className={cn("relative z-0 flex items-center gap-1 border-b", className)}
      {...props}
    >
      {children}
      <TabsPrimitive.Indicator className="bg-foreground absolute bottom-0 left-(--active-tab-left) -z-1 h-px w-(--active-tab-width) transition-all duration-200 ease-out" />
    </TabsPrimitive.List>
  );
}

function TabsTab({ className, ...props }: TabsPrimitive.Tab.Props) {
  return (
    <TabsPrimitive.Tab
      data-slot="tabs-tab"
      className={cn(
        "text-muted-foreground has-hover:hover:text-foreground data-active:text-foreground focus-visible:ring-primary/8-10 flex h-9 items-center gap-1.5 rounded-md px-2 text-sm font-medium whitespace-nowrap outline-hidden select-none focus-visible:ring-1 [&_svg]:size-4 [&_svg]:shrink-0",
        className,
      )}
      {...props}
    />
  );
}

function TabsPanel({ className, ...props }: TabsPrimitive.Panel.Props) {
  return (
    <TabsPrimitive.Panel
      data-slot="tabs-panel"
      className={cn("outline-hidden", className)}
      {...props}
    />
  );
}

export { Tabs, TabsList, TabsTab, TabsPanel };
