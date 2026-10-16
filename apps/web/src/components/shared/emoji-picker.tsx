import EmojiPickerReact, { EmojiStyle, Theme } from "emoji-picker-react";

type TProps = {
  theme: "light" | "dark";
  onSelect: (emoji: string) => void;
};

export default function EmojiPicker({ theme, onSelect }: TProps) {
  return (
    <EmojiPickerReact
      width="100%"
      height={320}
      theme={theme === "light" ? Theme.LIGHT : Theme.DARK}
      emojiStyle={EmojiStyle.NATIVE}
      lazyLoadEmojis
      previewConfig={{ showPreview: false }}
      skinTonesDisabled
      onEmojiClick={(data) => onSelect(data.emoji)}
    />
  );
}
