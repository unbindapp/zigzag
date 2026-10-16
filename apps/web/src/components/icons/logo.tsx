import { cn } from "@/components/ui/utils";

const markPaths = [
  "M 438 386 H 733 C 749 386 759 393 770 404 L 806 439 C 826 457 824 481 804 505 L 545 665 C 526 678 504 675 488 660 C 470 642 471 619 487 604 L 698 465 H 439 C 417 465 400 448 398 426 C 395 403 413 386 438 386 Z",
  "M 731 615 C 756 615 771 634 770 656 C 770 672 763 679 748 688 L 559 800 H 813 C 837 800 853 816 853 840 C 853 864 837 880 813 880 H 523 C 510 880 500 876 490 867 L 446 825 C 427 807 432 782 449 765 C 453 761 459 757 465 753 L 703 621 C 714 615 721 615 731 615 Z",
];

export function LogoMark({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 1254 1254" aria-hidden className={cn("size-5 shrink-0", className)}>
      <rect width="1254" height="1254" rx="240" className="fill-logo" />
      {markPaths.map((d) => (
        <path key={d} d={d} className="fill-logo-foreground" />
      ))}
    </svg>
  );
}

export function Logo() {
  return (
    <span className="inline-flex items-center gap-2">
      <LogoMark className="size-8" />
      <span className="text-xl leading-tight font-semibold">Zigzag</span>
    </span>
  );
}
