import { AvatarFallback, AvatarImage, AvatarRoot, type TAvatarSize } from "@/components/ui/avatar";
import { colorFromId, initials } from "@/lib/helpers/colors";

type TProps = {
  id: string;
  name: string;
  src?: string | null;
  size?: TAvatarSize;
};

export function Avatar({ id, name, src, size = "md" }: TProps) {
  return (
    <AvatarRoot size={size} title={name}>
      {src && <AvatarImage src={src} alt={name} />}
      <AvatarFallback className="text-white" style={{ backgroundColor: colorFromId(id) }}>
        {size === "xs" ? initials(name).slice(0, 1) : initials(name)}
      </AvatarFallback>
    </AvatarRoot>
  );
}
