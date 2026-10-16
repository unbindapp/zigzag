import ReactMarkdown, { defaultUrlTransform } from "react-markdown";
import remarkGfm from "remark-gfm";

import { mentionToLink } from "@/components/shared/markdown-mention";

type TProps = {
  markdown: string;
};

function urlTransform(url: string) {
  if (url.startsWith("user:")) return url;
  return defaultUrlTransform(url);
}

export function MarkdownView({ markdown }: TProps) {
  return (
    <div data-slot="markdown-view" className="zz-prose w-full">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        urlTransform={urlTransform}
        components={{
          a: ({ href, children }) => {
            if (href?.startsWith("user:")) {
              return <span data-type="mention">{children}</span>;
            }
            return (
              <a href={href} target="_blank" rel="noreferrer noopener">
                {children}
              </a>
            );
          },
        }}
      >
        {mentionToLink(markdown)}
      </ReactMarkdown>
    </div>
  );
}
