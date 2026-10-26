import { mediaUrl } from '@/lib/api';

export function removeMarkdownImage(content: string, imageUrl: string): string {
  const targetUrl = mediaUrl(imageUrl);

  return content
    .split('\n')
    .filter((line) => {
      const match = line.match(/!\[[^\]]*]\(([^)]+)\)/);
      return !match || mediaUrl(match[1]) !== targetUrl;
    })
    .join('\n');
}
