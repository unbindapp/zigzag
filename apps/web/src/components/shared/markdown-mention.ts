import { Mention } from "@tiptap/extension-mention";

const mentionPattern = /^@\[([^\]]+)\]\(user:([^)\s]+)\)/;
const mentionStartPattern = /@\[[^\]]+\]\(user:[^)\s]+\)/;

export const MarkdownMention = Mention.extend({
  parseMarkdown: (token, helpers) =>
    helpers.createNode("mention", { id: token.attributes?.id, label: token.attributes?.label }),
  markdownTokenizer: {
    name: "mention",
    level: "inline",
    start: (src) => src.search(mentionStartPattern),
    tokenize: (src) => {
      const match = src.match(mentionPattern);
      if (!match) return undefined;
      return { type: "mention", raw: match[0], attributes: { label: match[1], id: match[2] } };
    },
  },
  renderMarkdown: (node) => `@[${node.attrs?.label ?? ""}](user:${node.attrs?.id ?? ""})`,
});

export function mentionToLink(markdown: string) {
  return markdown.replace(/@\[([^\]]+)\]\(user:([^)\s]+)\)/g, "[@$1](user:$2)");
}
