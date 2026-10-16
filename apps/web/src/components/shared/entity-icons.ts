import {
  BeakerIcon,
  BookOpenIcon,
  BoxIcon,
  BrainIcon,
  BriefcaseIcon,
  BugIcon,
  CloudIcon,
  CodeIcon,
  CompassIcon,
  CpuIcon,
  DatabaseIcon,
  FlagIcon,
  FlameIcon,
  GemIcon,
  GlobeIcon,
  HeartIcon,
  LayersIcon,
  LightbulbIcon,
  type LucideIcon,
  MegaphoneIcon,
  PaletteIcon,
  PuzzleIcon,
  RocketIcon,
  ServerIcon,
  ShieldIcon,
  SmartphoneIcon,
  SparklesIcon,
  StarIcon,
  TargetIcon,
  TrophyIcon,
  UsersIcon,
  WrenchIcon,
  ZapIcon,
} from "lucide-react";

export const ENTITY_ICONS: Record<string, LucideIcon> = {
  rocket: RocketIcon,
  box: BoxIcon,
  layers: LayersIcon,
  code: CodeIcon,
  bug: BugIcon,
  zap: ZapIcon,
  target: TargetIcon,
  flag: FlagIcon,
  star: StarIcon,
  heart: HeartIcon,
  flame: FlameIcon,
  sparkles: SparklesIcon,
  lightbulb: LightbulbIcon,
  compass: CompassIcon,
  globe: GlobeIcon,
  cloud: CloudIcon,
  server: ServerIcon,
  database: DatabaseIcon,
  cpu: CpuIcon,
  smartphone: SmartphoneIcon,
  shield: ShieldIcon,
  wrench: WrenchIcon,
  puzzle: PuzzleIcon,
  palette: PaletteIcon,
  megaphone: MegaphoneIcon,
  briefcase: BriefcaseIcon,
  book: BookOpenIcon,
  beaker: BeakerIcon,
  brain: BrainIcon,
  gem: GemIcon,
  trophy: TrophyIcon,
  users: UsersIcon,
};

export const ENTITY_ICON_NAMES = Object.keys(ENTITY_ICONS);

export type TResolvedIcon =
  { kind: "lucide"; Icon: LucideIcon } | { kind: "emoji"; emoji: string } | { kind: "none" };

const ICON_NAME = /^[A-Za-z][\w-]*$/;

export function resolveEntityIcon(icon: string | null | undefined): TResolvedIcon {
  if (!icon) return { kind: "none" };
  if (Object.hasOwn(ENTITY_ICONS, icon)) return { kind: "lucide", Icon: ENTITY_ICONS[icon] };
  if (ICON_NAME.test(icon)) return { kind: "none" };
  return { kind: "emoji", emoji: icon };
}
