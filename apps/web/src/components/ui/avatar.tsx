import { Avatar as AvatarPrimitive } from "@base-ui/react/avatar";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/components/ui/utils";

const avatarVariants = cva(
  "bg-muted text-foreground relative inline-flex shrink-0 items-center justify-center overflow-hidden rounded-full align-middle font-medium select-none",
  {
    variants: {
      size: {
        xs: "size-4 text-[0.5rem]",
        sm: "size-5 text-[0.5625rem]",
        md: "size-6 text-[0.625rem]",
        lg: "size-8 text-xs",
        xl: "size-16 text-xl",
      },
    },
    defaultVariants: { size: "md" },
  },
);

export type TAvatarSize = NonNullable<VariantProps<typeof avatarVariants>["size"]>;

function AvatarRoot({
  className,
  size,
  ...props
}: AvatarPrimitive.Root.Props & VariantProps<typeof avatarVariants>) {
  return (
    <AvatarPrimitive.Root
      data-slot="avatar"
      className={cn(avatarVariants({ size }), className)}
      {...props}
    />
  );
}

function AvatarImage({ className, ...props }: AvatarPrimitive.Image.Props) {
  return (
    <AvatarPrimitive.Image
      data-slot="avatar-image"
      className={cn("size-full object-cover", className)}
      {...props}
    />
  );
}

function AvatarFallback({ className, ...props }: AvatarPrimitive.Fallback.Props) {
  return (
    <AvatarPrimitive.Fallback
      data-slot="avatar-fallback"
      className={cn("flex size-full items-center justify-center leading-none", className)}
      {...props}
    />
  );
}

export { AvatarRoot, AvatarImage, AvatarFallback, avatarVariants };
